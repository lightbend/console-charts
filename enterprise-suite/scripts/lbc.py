#!/usr/bin/env python

from __future__ import print_function
from __future__ import unicode_literals

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
import zipfile
import datetime
import base64
from distutils.version import LooseVersion

# Note: this script has to run on both python2 and python3, because we dont know
# which version 'python' will be installed on a host system. The main difference
# to keep in mind is 'bytes' vs 'str' types. Following are some helper methods.

def is_python2():
    return sys.version_info < (3, 0)

if is_python2():
    import urllib2 as url
else:
    import urllib.request as url

def bytes_to_str(s):
    if is_python2():
        return s
    else:
        return str(s, encoding='utf-8')

def str_to_bytes(s):
    if is_python2():
        return s
    else:
        return s.encode('utf-8')

# Required dependency versions
REQ_VER_KUBECTL = '1.10'
REQ_VER_HELM = '2.10'
REQ_VER_MINIKUBE = '0.29'
REQ_VER_MINISHIFT = '1.20'
REQ_VER_OC = '3.9'

DEFAULT_TIMEOUT=3

# Parsed commandline args
args = None

# The following functions are overridable for testing purposes

# Prints to stderr
def printerr(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

# Prints to stdout
def printinfo(*args, **kwargs):
    print(*args, **kwargs)

# Exits process with a message and non-0 exit code
def fail(msg):
    sys.exit(msg)

def make_tempdir():
    return tempfile.mkdtemp()

# Runs a given command with optional timeout.
# Returns (stdout, returncode) tuple. If timeout
# occured, returncode will be negative (-9 on macOS).
def run(cmd, timeout=None, stdin=None, show_stderr=True):
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
        if len(stderr) > 0 and show_stderr:
            printerr(bytes_to_str(stderr))
        returncode = proc.returncode
    finally:
        if timer != None:
            timer.cancel()
        return bytes_to_str(stdout), returncode

# Executes a command if dry_run=False,
# prints it to stdout or stderr, handles failure status
# codes by exiting with and error if can_fail=False.
def execute(cmd, can_fail=False, print_to_stdout=False):
    printerr(cmd)
    if not args.dry_run:
        stdout, returncode = run(cmd)
        if print_to_stdout:
            printinfo(stdout)
        else:
            printerr(stdout)
        if not can_fail and returncode != 0:
            fail("Command '" + cmd + "' failed!")
        return returncode
    return 0

version_re = re.compile(r'([0-9]+\.[0-9]+(\.[0-9]+)?)')
def require_version(cmd, required_version):
    # Use first word as a program name in messages
    name = cmd.partition(' ')[0]

    # Use 1s timeout, mainly for docker when DOCKER_HOST is unreachable
    stdout, returncode = run(cmd, 1)

    if returncode == None:
        fail("Required program '" + name + "' not found")
    elif returncode == 0 and stdout != '':
        match = version_re.search(stdout)
        if match != None:
            current = LooseVersion(match.group())
            required = LooseVersion(required_version)
            if current >= required:
                return
            else:
                fail("Installed version of '" + name + "' is too old. Found: {}, required: {}"
                     .format(current, required)) 

    # Non-critical warning
    printerr("warning: unable to determine installed version of '" + name + "'")

def is_running_minikube():
    stdout, returncode = run('minikube status')
    if returncode == 0:
        if ('minikube: Running' in stdout) and ('cluster: Running') in stdout:
            stdout, returncode = run('kubectl config current-context')
            return returncode == 0 and stdout == 'minikube'
    return False

def is_running_minishift():
    stdout, returncode = run('minishift status')
    if returncode == 0:
        if ('minishift: Running' in stdout) and ('cluster: Running') in stdout:
            stdout, returncode = run('kubectl config current-context')
            return returncode == 0 and stdout == 'minishift'
    return False

# Helm check is a separate function because we also need it when not doing full
# preflight check, eg. when using --export-yaml argument
def check_helm():
    require_version('helm version --client --short', REQ_VER_HELM)

# Kubectl check is needed both in install and verify subcommands
def check_kubectl(minishift=False):
    require_version('kubectl version --client=true --short=true', REQ_VER_KUBECTL)

    # Check if kubectl is connected to a cluster. If not connected, version query will timeout.
    stdout, returncode = run('kubectl version', DEFAULT_TIMEOUT)
    if returncode != 0:
        msg = 'Cannot reach cluster with kubectl'
        if minishift:
            # Minishift needs special configuration for kubectl to work
            msg = msg + ". Did you do 'eval $(minishift oc-env)'?"
        fail(msg)

def check_credentials(creds):
    registry = 'https://lightbend-docker-commercial-registry.bintray.io/v2'
    api_url = registry + '/enterprise-suite/es-monitor-api/tags/list'

    # Set up basic auth with given creds
    req = url.Request(api_url)
    basic_auth = base64.b64encode(str_to_bytes('{}:{}'.format(creds[0], creds[1])))
    req.add_header('Authorization', 'Basic ' + bytes_to_str(basic_auth))

    success = False
    try:
        resp = url.urlopen(req)
        if resp.getcode() == 200:
            # Lazy way of verifying returned json - there should be a tag named "latest"
            if '"latest"' in bytes_to_str(resp.read()):
                success = True
    finally:
        return success 

def preinstall_check(creds, minikube=False, minishift=False):
    check_helm()
    check_kubectl()

    if minikube:
        require_version('minikube version', REQ_VER_MINIKUBE)

    if minishift:
        require_version('minishift version', REQ_VER_MINISHIFT)
        require_version('oc version', REQ_VER_OC)

    # Check if helm is set up inside a cluster
    stdout, returncode = run('helm version', DEFAULT_TIMEOUT)
    if returncode != 0:
        fail('Cannot get helm status. Did you set up helm inside your cluster?')

    # TODO: Check if RBAC rules for tiller are set up

    if not check_credentials(creds):
        fail('Your credentials do not appear to be correct' +
                 ' - unable to make authenticated request to lightbend docker registry')

def install_helm_chart(creds_file):
    creds_arg = '--values ' + creds_file
    # Helm args are separated from lbc.py args by double dash, filter it out
    helm_args = ' '.join([arg for arg in args.helm if arg != '--'])
    version_arg = ('--version ' + args.version) if args.version != None else '--devel'

    chart_ref = None
    if args.local_chart != None:
        # Install from local chart tarball
        chart_ref = args.local_chart
    else:
        execute('helm repo add es-repo ' + args.repo)
        execute('helm repo update')
        chart_ref = 'es-repo/' + args.chart
    
    if args.export_yaml != None:
        # Tilerless path - renders kubernetes resources and prints to stdout

        creds_exec_arg = ''
        if args.export_yaml == 'creds':
            creds_exec_arg = '--execute templates/commercial-credentials.yaml ' + creds_arg
            printerr('warning: credentials in yaml are not encrypted, only base64 encoded. Handle appropriately.')
        
        try:
            chartfile = args.local_chart
            tempdir = None
            if chartfile == None:
                # No local chart given, fetch from repo
                tempdir = make_tempdir()
                execute('helm fetch -d {} {} {}'
                    .format(tempdir, version_arg, chart_ref))
                chartfile_glob = tempdir + '/' + args.chart + '*.tgz'
                # Print a fake chart archive name when dry-running
                chartfile = glob.glob(chartfile_glob) if not args.dry_run else ['enterprise-suite-ver.tgz']
                if len(chartfile) < 1: 
                    fail('cannot access fetched chartfile at {}, ES_CHART={}'
                        .format(chartfile_glob, args.chart))
                chartfile = chartfile[0]
            execute('helm template --name {} --namespace {} {} {} {}'
                .format(args.helm_name, args.namespace, helm_args,
                creds_exec_arg, chartfile), print_to_stdout=True)
        finally:
            if tempdir != None:
                shutil.rmtree(tempdir)

    else:
        # Tiller path - installs console directly to a k8s cluster in a given namespace

        # Determine if we should upgrade or install
        should_upgrade = False
        stdout, returncode = run('helm status ' + args.helm_name,
                                 DEFAULT_TIMEOUT, show_stderr=False)
        if returncode == 0:
            if args.force_install:
                execute('helm delete --purge ' + args.helm_name)
                printerr(('warning: helm delete does not wait for resources to be removed'
                          '- if the script fails on install, please re-run it.'))
            else:
                should_upgrade = True
    
        if should_upgrade:
            execute('helm upgrade {} {} {} {} {}'
                .format(args.helm_name, chart_ref, version_arg,
                        creds_arg, helm_args))
        else:
            execute('helm install {} --name {} --namespace {} {} {} {}'
                .format(chart_ref, args.helm_name, args.namespace,
                        version_arg, creds_arg, helm_args))

def write_temp_credentials(creds_tempfile, creds):
    creds_str = '\n'.join(["imageCredentials:",
                           "    username: " + creds[0],
                           "    password: " + creds[1]])
    creds_tempfile.write(creds_str)
    creds_tempfile.flush()

def import_credentials():
    creds = (os.environ.get('LIGHTBEND_COMMERCIAL_USERNAME'),
             os.environ.get('LIGHTBEND_COMMERCIAL_PASSWORD'))

    if creds[0] == None or creds[1] == None:
        with open(os.path.expanduser(args.creds), 'r') as creds_file:
            creds_dict = dict(re.findall(r'(\S+)\s*=\s*(".*?"|\S+)', creds_file.read()))
            creds = (creds_dict.get('user'), creds_dict.get('password'))

    if creds[0] == None or creds[1] == None:
        fail("Credentials missing, please check your credentials file\n"
             "LIGHTBEND_COMMERCIAL_CREDENTIALS=" + args.creds)

    return creds

def check_install():
    def deployment_running(name):
        printinfo('Checking deployment {} ... '.format(name), end='')
        stdout, returncode = run('kubectl --namespace {} get deploy/{} --no-headers'
                                 .format(args.namespace, name))
        if returncode == 0:
            # Skip first (deployment name) and last (running time) items
            cols = [int(col) for col in stdout.split()[1:-1]]
            desired, current, up_to_date, available = cols[0], cols[1], cols[2], cols[3]
            if desired <= 0:
                printinfo('failed')
                printerr('Deployment {} status check: expected to see 1 or more desired replicas, found 0'
                         .format(name))
            if desired > available:
                printinfo('failed')
                printerr('Deployment {} status check: available replica number ({}) is less than desired ({})'
                         .format(name, available, desired))
            if desired > 0 and desired == available:
                printinfo('ok')
                return True
        else:
            printinfo('failed')
            printerr('Unable to check deployment {} status'.format(name))
        return False

    status_ok = True
    status_ok &= deployment_running('es-console')
    status_ok &= deployment_running('grafana-server')
    status_ok &= deployment_running('prometheus-alertmanager')
    status_ok &= deployment_running('prometheus-kube-state-metrics')
    status_ok &= deployment_running('prometheus-server')

    if status_ok:
        printerr('Your Lightbend Console seems to be running fine!')
    else:
        fail('Lightbend Console status check failed')

def debug_dump(args):

    def dump(dest, filename, content):
        if args.print:
            # Print to stdout
            printinfo('=== File: {} ==='.format(filename))
            printinfo(content)
        else:
            # Put to a zipfile
            dest.writestr(filename, stdout)

    def get_pod_containers(pod):
        # This magic gives us all the containers in a pod
        containers, returncode = run("kubectl --namespace {} get pods {} -o jsonpath='{{.spec.containers[*].name}}'"
                                     .format(args.namespace, pod))
        if returncode == 0:
            return containers.split()
        else:
            fail(failure_msg + 'unable to get containers in a pod {}'
                 .format(pod))

    def write_log(archive, pod, container):
        stdout, returncode = run('kubectl --namespace {} logs {} -c {}'
                                 .format(args.namespace, pod, container),
                                 show_stderr=False)
        if returncode == 0:
            filename = '{}+{}.log'.format(pod, container)
            dump(archive, filename, stdout)
        else:
            fail(failure_msg + 'unable to get logs for container {} in a pod {}'
                     .format(container, pod))

        # Try to get previous logs too
        stdout, returncode = run('kubectl --namespace {} logs {} -c {} -p'
                                 .format(args.namespace, pod, container),
                                 show_stderr=False)
        if returncode == 0:
            filename = '{}+{}+prev.log'.format(pod, container)
            dump(archive, filename, stdout)

    failure_msg = 'Failed to get diagnostic data: '
    
    archive = None
    if not args.print:
        timestamp = datetime.datetime.now().strftime('%Y-%m-%d-%H-%M-%S')
        filename = 'console-diagnostics-{}.zip'.format(timestamp)
        archive = zipfile.ZipFile(filename, 'w')

    # List all resources in our namespace
    stdout, returncode = run('kubectl --namespace {} get all'.format(args.namespace),
                             show_stderr=False)
    if returncode == 0:
        dump(archive, 'kubectl-get-all.txt', stdout)
    else:
        fail(failure_msg + 'unable to list k8s resources in {} namespace'
                .format(args.namespace))

    # Describe all resources
    stdout, returncode = run('kubectl --namespace {} describe all'.format(args.namespace),
                             show_stderr=False)
    if returncode == 0:
        dump(archive, 'kubectl-describe-all.txt', stdout)
    else:
        fail(failure_msg + 'unable to describe k8s resources in {} namespace'
                .format(args.namespace))

    # Iterate over pods
    stdout, returncode = run('kubectl --namespace {} get pods --no-headers'.format(args.namespace))
    if returncode == 0:
        for line in stdout.split('\n'):
            if len(line) > 0:
                # Pod name is the first thing on the line
                pod = line.split()[0]

                for container in get_pod_containers(pod):
                    write_log(archive, pod, container)
    else:
        fail(failure_msg + 'unable to pods in the namespace {}'.format(args.namespace))

    if archive != None:
        archive.close()
        printerr('Lightbend Console diagnostic data written to {}'.format(filename))

def setup_args(argv):
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest='subcommand', help='sub-command help')

    fmt = argparse.ArgumentDefaultsHelpFormatter
    install = subparsers.add_parser('install', help='install lightbend console', formatter_class=fmt)
    verify = subparsers.add_parser('verify', help='verify console installation', formatter_class=fmt)
    debug_dump = subparsers.add_parser('debug-dump', help='make an archive with k8s status info for debugging and diagnostic purposes',
                                       formatter_class=fmt)

    # Debug dump arguments
    debug_dump.add_argument('--print', help='print the output instead of writing it to a zip file',
                        action='store_true')

    # Install arguments
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
    install.add_argument('--creds', help='credentials file', default='~/.lightbend/commercial.credentials')
    install.add_argument('--version', help='console version to install', type=str)

    install.add_argument('helm', help="any additional arguments separated by '--' will be passed to helm (eg. '-- --set emptyDir=false')",
                         nargs=argparse.REMAINDER)

    # Common arguments for all subparsers
    for subparser in [install, verify, debug_dump]:
        subparser.add_argument('--skip-checks', help='skip environment checks',
                               action='store_true')
        subparser.add_argument('--namespace', help='namespace to install console into/where it is installed', default='lightbend')

    args = parser.parse_args(argv)

    if len(argv) == 0:
        parser.print_help()

    return args

def main(argv):
    global args
    args = setup_args(argv)

    if args.subcommand == 'verify':
        if not args.skip_checks:
            check_kubectl()
        check_install()
    
    if args.subcommand == 'install':
        creds = import_credentials()

        if not args.skip_checks:
            if args.export_yaml == None:
                minikube = is_running_minikube()
                minishift = is_running_minishift()
                preinstall_check(creds, minikube, minishift)
            else:
                check_helm()

        if args.version == None:
            printerr(("warning: --version has not been set, helm will use the latest available version. "
                "It is recommended to use an explicit version."))

        with tempfile.NamedTemporaryFile('w') as creds_tempfile:
            write_temp_credentials(creds_tempfile, creds)
            install_helm_chart(creds_tempfile.name)

    if args.subcommand == 'debug-dump':
        if not args.skip_checks:
            check_kubectl()
        debug_dump(args)

if __name__ == '__main__':
    main(sys.argv[1:])
