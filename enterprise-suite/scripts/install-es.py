#!/usr/bin/env python

import sys 
import os
import shlex
import subprocess
import threading
import tempfile
import re
import argparse
from distutils.version import LooseVersion

DEFAULT_TIMEOUT=3

# Parsed commandline args
args = None

# Runs a given command with optional timeout.
# Returns (stdout, returncode) tuple. If timeout
# occured, returncode will be negative (-9 on macOS).
def run(cmd, timeout=None, stdin=None):
    stdout, stderr, returncode = None, None, None
    try:
        proc = subprocess.Popen(shlex.split(cmd),
                                stdout=subprocess.PIPE,
                                stdin=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        timer = threading.Timer(timeout, proc.kill) if timeout != None else None 
        if timer != None:
            timer.start()
        stdout, stderr = proc.communicate(input=stdin)
        returncode = proc.returncode
    finally:
        if timer != None:
            timer.cancel()
        return stdout, returncode

# Executes a command if dry_run=False,
# prints it to stdout, handles failure status
# codes by exiting with and error if can_fail=False.
def execute(cmd, can_fail=False):
    print(cmd)
    if not args.dry_run:
        stdout, returncode = run(cmd)
        if not can_fail and returncode != 0:
            sys.exit("Command '" + cmd + "' failed:\n" + stdout)
        return returncode
    return 0

version_re = re.compile(r'([0-9]+\.[0-9]+(\.[0-9]+)?)')
def require_version(cmd, required_version):
    # Use first word as a program name in messages
    name = cmd.partition(' ')[0]

    # Use 1s timeout, mainly for docker when DOCKER_HOST is unreachable
    stdout, returncode = run(cmd, 1)

    if returncode == None:
        sys.exit("Required program '" + name + "' not found")
    elif returncode == 0 and stdout != '':
        match = version_re.search(stdout)
        if match != None:
            current = LooseVersion(match.group())
            required = LooseVersion(required_version)
            if current >= required:
                return
            else:
                sys.exit("Installed version of '" + name + "' is too old. Found: {}, required: {}"
                    .format(current, required)) 

    # Non-critical warning
    print("warning: unable to determine installed version of '" + name + "'")

def preflight_check(creds, minikube=False):
    # Check versions
    require_version("docker version -f '{{.Client.Version}}'", '17.06')
    require_version('kubectl version --client=true --short=true', '1.10')
    require_version('helm version --client --short', '2.10')
    if minikube:
        require_version('minikube version', '0.29')

    # Check if kubectl is connected to a cluster. If not connected, version query will timeout.
    stdout, returncode = run('kubectl version', DEFAULT_TIMEOUT)
    if returncode != 0:
        sys.exit('Cannot reach cluster with kubectl')

    if minikube:
        # Check if docker is pointing to a cluster
        if os.environ.get('DOCKER_HOST') == None:
            sys.exit('Docker CLI is not pointing to a cluster. Did you run "eval $(minikube docker-env)"?')

    # Check if helm is set up inside a cluster
    stdout, returncode = run('helm version', DEFAULT_TIMEOUT)
    if returncode != 0:
        sys.exit('Cannot get helm status. Did you set up helm inside your cluster?')

    # TODO: Check if RBAC rules for tiller are set up

    # Check credentials
    registry = 'lightbend-docker-commercial-registry.bintray.io'
    stdout, returncode = run('docker login -u {} --password-stdin {}'.format(creds[0], registry),
                             6, creds[1])
    if returncode != 0:
        print(stdout)
        sys.exit('Unable to login to lightbend docker registry. Please check your credentials.')
    else:
        run('docker logout ' + registry, DEFAULT_TIMEOUT)
    
    # TODO: Try to pull docker image from lighbend registry
    
    # At the moment when I try 'docker pull {registry}/enterprise-suite/es-monitor-api:latest' I get:
    # Error response from daemon: Get https://lightbend-commercial-registry.bintray.io/v2/: 
    # x509: certificate is valid for *.bintray.com, bintray.com, not lightbend-commercial-registry.bintray.io

def install_helm_chart(args, creds_file):
    chart_ref = None
    if args.local_chart != None:
        # Install from local chart tarball
        chart_ref = args.local_chart
    else:
        execute('helm repo add es-repo ' + args.repo)
        execute('helm repo update')
        chart_ref = 'es-repo/' + args.chart
    
    if args.export_yaml != 'false':
        # TODO
        pass
    else:
        # Determine if we should upgrade or install
        should_upgrade = False
        stdout, returncode = run('helm status ' + args.helm_name, DEFAULT_TIMEOUT)
        if returncode == 0:
            if args.force_install:
                execute('helm delete --purge ' + args.helm_name)
                print(('warning: helm delete does not wait for resources to be removed'
                       '- if the script fails on install, please re-run it.'))
            else:
                should_upgrade = True
    
        # TODO: import credentials

        # TODO: add credentials, --set args
        if should_upgrade:
            execute('helm upgrade {} {}'
                .format(args.helm_name, chart_ref))
        else:
            execute('helm install {} --name {}'
                .format(chart_ref, args.helm_name))

def write_temp_credentials(creds_tempfile, creds):
    creds_tempfile.write("imageCredentials" +
                    "    username: " + creds[0] +
                    "    password: " + creds[1])
    creds_tempfile.flush()

def import_credentials(args):
    creds = (os.environ.get('LIGHTBEND_COMMERCIAL_USERNAME'),
             os.environ.get('LIGHTBEND_COMMERCIAL_PASSWORD'))

    if creds[0] == None or creds[1] == None:
        with open(os.path.expanduser(args.creds), 'r') as creds_file:
            creds_dict = dict(re.findall(r'(\S+)\s*=\s*(".*?"|\S+)', creds_file.read()))
            creds = (creds_dict.get('user'), creds_dict.get('password'))

    if creds[0] == None or creds[1] == None:
        sys.exit("Credentials missing, please check your credentials file\n"
                 "LIGHTBEND_COMMERCIAL_CREDENTIALS=" + args.creds)

    return creds

def setup_args():
    parser = argparse.ArgumentParser()

    parser.add_argument('--dry-run', help='only print out the commands that will be executed',
                        action='store_true')
    parser.add_argument('--force-install', help='set to true to delete an existing install first, instead of upgrading',
                        action='store_true')
    parser.add_argument('--export-yaml', help='export resource yaml to stdout, forces dry_run if not false',
                        choices=['creds', 'console', 'false'], default='false')
    parser.add_argument('--helm-name', help='helm release name', default='enterprise-suite')
    parser.add_argument('--local-chart', help='set to location of local chart tarball')
    parser.add_argument('--namespace', help='namespace to install es-console into', default='lightbend')
    parser.add_argument('--chart', help='chart name to install from the repository', default='enterprise-suite')
    parser.add_argument('--repo', help='helm chart repository', default='https://repo.lightbend.com/helm-charts')
    parser.add_argument('--creds', help='credentials file', default='~/.lightbend/commercial.credentials')
    parser.add_argument('--version', help='es-console version to install', type=str)

    return parser.parse_args()

def main():
    global args
    args = setup_args()
    creds = import_credentials(args)

    preflight_check(creds)

    if args.version == None:
        print(("warning: --version has not been set, helm will use the latest available version. "
               "It is recommended to use an explicit version."))

    with tempfile.NamedTemporaryFile('w') as creds_tempfile:
        write_temp_credentials(creds_tempfile, creds)
        install_helm_chart(args, creds)

if __name__ == '__main__':
    main()