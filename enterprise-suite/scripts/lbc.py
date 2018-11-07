#!/usr/bin/env python

from __future__ import print_function

import sys 
import os
import glob
import shlex
import shutil
import subprocess
import threading
import tempfile
import re
import argparse
from distutils.version import LooseVersion

# Required versions
REQ_VER_DOCKER = '17.06'
REQ_VER_KUBECTL = '1.10'
REQ_VER_HELM = '2.10'
REQ_VER_MINIKUBE = '0.29'
REQ_VER_MINISHIFT = '1.20'

DEFAULT_TIMEOUT=3

# Parsed commandline args
args = None

# Prints to stderr
def printerr(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

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
        if len(stderr) > 0:
            printerr(stderr)
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
        print(stdout)
        if not can_fail and returncode != 0:
            sys.exit("Command '" + cmd + "' failed!")
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
    printerr("warning: unable to determine installed version of '" + name + "'")

# Helm check is a separate function because we also need it when not doing full
# preflight check, eg. when using --export-yaml argument
def check_helm():
    require_version('helm version --client --short', REQ_VER_HELM)

def preflight_check(creds, minikube=False):
    # Check versions
    require_version("docker version -f '{{.Client.Version}}'", REQ_VER_DOCKER)
    require_version('kubectl version --client=true --short=true', REQ_VER_KUBECTL)
    check_helm()
    if minikube:
        require_version('minikube version', REQ_VER_MINIKUBE)

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
        sys.exit('Unable to login to lightbend docker registry. Please check your credentials.')
    else:
        run('docker logout ' + registry, DEFAULT_TIMEOUT)
    
    # TODO: Try to pull docker image from lighbend registry
    
    # At the moment when I try 'docker pull {registry}/enterprise-suite/es-monitor-api:latest' I get:
    # Error response from daemon: Get https://lightbend-commercial-registry.bintray.io/v2/: 
    # x509: certificate is valid for *.bintray.com, bintray.com, not lightbend-commercial-registry.bintray.io

def install_helm_chart(args, creds_file):
    creds_values = '--values ' + creds_file
    helm_args = ' '.join([arg for arg in args.helm if arg != '--'])

    chart_ref = None
    if args.local_chart != None:
        # Install from local chart tarball
        chart_ref = args.local_chart
    else:
        execute('helm repo add es-repo ' + args.repo)
        execute('helm repo update')
        chart_ref = 'es-repo/' + args.chart
    
    if args.export_yaml != 'false':
        credentials_arg = ''
        if args.export_yaml == 'creds':
            credentials_arg = '--execute templates/commercial-credentials.yaml ' + creds_values
            printerr('warning: credentials in yaml are not encrypted, only base64 encoded. Handle appropriately.')
        
        try:
            tempdir = tempfile.mkdtemp()
            execute('helm fetch -d {} {} {}'
                .format(tempdir, args.version or '', chart_ref))
            chartfile_glob = tempdir + '/' + args.chart + '*.tgz'
            chartfile = glob.glob(chartfile_glob) if not args.dry_run else ['enterprise-suite-latest.tgz']
            if len(chartfile) < 1: 
                sys.exit('cannot access fetched chartfile at {}, ES_CHART={}'
                    .format(chartfile_glob, args.chart))
            execute('helm template --name {} --namespace {} {} {} {}'
                .format(args.helm_name, args.namespace, helm_args,
                credentials_arg, chartfile[0]))
        finally:
            shutil.rmtree(tempdir)

    else:
        # Determine if we should upgrade or install
        should_upgrade = False
        stdout, returncode = run('helm status ' + args.helm_name, DEFAULT_TIMEOUT)
        if returncode == 0:
            if args.force_install:
                execute('helm delete --purge ' + args.helm_name)
                printerr(('warning: helm delete does not wait for resources to be removed'
                       '- if the script fails on install, please re-run it.'))
            else:
                should_upgrade = True
    
        if should_upgrade:
            execute('helm upgrade {} {} {} {}'
                .format(args.helm_name, chart_ref, creds_values,
                        helm_args))
        else:
            execute('helm install {} --name {} --namespace {} {} {}'
                .format(chart_ref, args.helm_name, args.namespace,
                        creds_values, helm_args))

def write_temp_credentials(creds_tempfile, creds):
    #creds_tempfile.write("imageCredentials" +
    credsstr = '\n'.join(["imageCredentials:",
                          "    username: " + creds[0],
                          "    password: " + creds[1]])
    creds_tempfile.write(credsstr)
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
    subparsers = parser.add_subparsers(dest='subcommand', help='sub-command help')

    install = subparsers.add_parser('install', help='install es-console')
    verify = subparsers.add_parser('verify', help='verify es-console installation')
    diagnose = subparsers.add_parser('diagnose', help='make an archive with k8s logs for debugging and diagnostic purposes')

    parser.add_argument('--creds', help='credentials file', default='~/.lightbend/commercial.credentials')
    parser.add_argument('--namespace', help='namespace to install es-console into/where it is installed', default='lightbend')

    install.add_argument('--dry-run', help='only print out the commands that will be executed',
                        action='store_true')
    install.add_argument('--force-install', help='set to true to delete an existing install first, instead of upgrading',
                        action='store_true')
    install.add_argument('--export-yaml', help='export resource yaml to stdout',
                        choices=['creds', 'console'])
    install.add_argument('--helm-name', help='helm release name', default='enterprise-suite')
    install.add_argument('--local-chart', help='set to location of local chart tarball')
    install.add_argument('--chart', help='chart name to install from the repository', default='enterprise-suite')
    install.add_argument('--repo', help='helm chart repository', default='https://repo.lightbend.com/helm-charts')
    install.add_argument('--version', help='es-console version to install', type=str)

    install.add_argument('helm', help='arguments to be passed to helm', nargs=argparse.REMAINDER)

    return parser.parse_args()

def main():
    global args
    args = setup_args()

    if args.subcommand == 'verify':
        creds = import_credentials(args)
        preflight_check(creds)
    
    if args.subcommand == 'install':
        creds = import_credentials(args)

        # TODO: autodetect minikube and minishift, do additional checks on them
        if args.export_yaml == None:
            preflight_check(creds)
        else:
            check_helm()

        if args.version == None:
            printerr(("warning: --version has not been set, helm will use the latest available version. "
                "It is recommended to use an explicit version."))

        with tempfile.NamedTemporaryFile('w') as creds_tempfile:
            write_temp_credentials(creds_tempfile, creds)
            install_helm_chart(args, creds_tempfile.name)

if __name__ == '__main__':
    main()