#!/usr/bin/env python

import sys 
import os
import shlex
import subprocess
import threading
import re
import argparse
from distutils.version import LooseVersion

DEFAULT_TIMEOUT=3

# Parsed commandline args
args = None

# Runs a given command with optional timeout.
# Returns (stdout, returncode) tuple. If timeout
# occured, returncode will be negative (-9 on macOS).
def run(cmd, timeout=None):
    stdout, stderr, returncode = None, None, None
    try:
        proc = subprocess.Popen(shlex.split(cmd),
                                stdout=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        timer = threading.Timer(timeout, proc.kill) if timeout != None else None 
        if timer != None:
            timer.start()
        stdout, stderr = proc.communicate()
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

def preflight_check():
    # Check versions
    require_version("docker version -f '{{.Client.Version}}'", '17.06')
    require_version('kubectl version --client=true --short=true', '1.10')
    require_version('helm version --client --short', '2.10')
    require_version('minikube version', '0.29')

    # Check if kubectl is connected to a cluster. If not connected, version query will timeout.
    stdout, returncode = run('kubectl version', DEFAULT_TIMEOUT)
    if returncode != 0:
        sys.exit('Cannot reach cluster with kubectl')

    # Check if docker is pointing to a cluster
    if os.environ.get('DOCKER_HOST') == None:
        sys.exit('Docker CLI is not pointing to a cluster. Did you run "eval $(minikube docker-env)"?')

    # Check if helm is set up inside a cluster
    stdout, returncode = run('helm version', DEFAULT_TIMEOUT)
    if returncode != 0:
        sys.exit('Cannot get helm status. Did you set up helm inside your cluster?')

    # TODO: Check if RBAC rules for tiller are set up

def install_helm_chart(args):
    chart_ref = None
    if args.local_chart != None:
        # Install from local chart tarball
        chart_ref = args.local_chart
    else:
        execute('helm repo add es-repo ' + es_repo)
        execute('helm repo update')
        chart_ref = 'es-repo/' + es_chart
    
    if args.export_yaml != 'false':
        # TODO
        pass
    else:
        # Determine if we should upgrade or install
        should_upgrade = False
        stdout, returncode = run('helm status ' + es_helm_name, DEFAULT_TIMEOUT)
        if returncode == 0:
            if args.force_install:
                execute('helm delete --purge ' + es_helm_name)
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
    print("DEBUG: args:" + str(args))

    preflight_check()

    if args.version == None:
        print(("warning: --version has not been set, helm will use the latest available version. "
               "It is recommended to use an explicit version."))
    
    install_helm_chart(args)

if __name__ == '__main__':
    main()