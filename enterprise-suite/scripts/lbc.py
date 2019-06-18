#!/usr/bin/env python

from __future__ import print_function

import sys
import subprocess

# Note: this script expects to run on python2. Some systems map 'python'
# executable name to python3, detect that and complain.
if sys.version_info >= (3, 0):
    proc = subprocess.run(['python2'] + sys.argv)
    sys.exit(proc.returncode)

import os
import shlex
import shutil
import threading
import tempfile
import re
import argparse
import zipfile
import datetime
import base64
import urllib2 as url
import json
import time
import glob
from distutils.version import LooseVersion

# Minimum required dependency versions
REQ_VER_KUBECTL = '1.10'
REQ_VER_HELM = '2.10'
REQ_VER_MINIKUBE = '0.29'
REQ_VER_MINISHIFT = '1.20'
REQ_VER_OC = '3.9'

# Verify looks for these deployments, need to be updated if helm chart changes!
CONSOLE_DEPLOYMENTS = [
    'console-backend',
    'console-frontend',
    'grafana',
    'prometheus-kube-state-metrics',
]

# Deployments for 1.1 and older versions of Console, install script needs to be compatible
CONSOLE_DEPLOYMENTS_OLD = [
    'es-console',
    'grafana-server',
    'prometheus-server',
    'prometheus-kube-state-metrics',
]

# Alertmanager deployment, this check can be turned off with --external-alertmanager
CONSOLE_ALERTMANAGER_DEPLOYMENT = 'prometheus-alertmanager'

# PVCs we need to pay attention to.
CONSOLE_PVCS = [
    'alertmanager-storage',
    'es-grafana-storage',
    'prometheus-storage'
]

DEFAULT_TIMEOUT = 10
REINSTALL_WAIT_SECS = 5

# Parsed commandline args
args = None

# Computed values
values = {}

# This will be true if running on Windows
windows = os.name == 'nt'

# The following functions are overridable for testing purposes


# Prints to stderr
def printerr(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)


# Prints to stdout
def printout(*args, **kwargs):
    print(*args, **kwargs)


# Exits process with a message and non-0 exit code
def fail(msg):
    sys.exit(msg)


# Runs a given command with optional timeout.
# Returns (returncode, stdout, stderr) tuple. If timeout
# occurred, returncode will be negative (-9 on macOS).
def run(cmd, timeout=None, stdin=None, show_stderr=True):
    stdout, stderr, returncode, timer = None, None, None, None
    try:
        proc = subprocess.Popen(shlex.split(cmd),
                                stdout=subprocess.PIPE,
                                stdin=subprocess.PIPE,
                                stderr=subprocess.PIPE)
        if timeout is not None:
            timer = threading.Timer(timeout, proc.kill)
            timer.start()
        stdout, stderr = proc.communicate(input=stdin)
        if len(stderr) > 0 and show_stderr:
            printerr(stderr)
        returncode = proc.returncode
    except OSError as e:
        stdout = e.strerror
        returncode = e.errno
    except Exception as e:
        stdout = str(e)
        returncode = 1
    finally:
        if timer is not None:
            timer.cancel()
        return returncode, stdout, stderr


# Executes a command if dry_run=False,
# prints it to stdout or stderr, handles failure status
# codes by exiting with an error if can_fail=False.
def execute(cmd, can_fail=False, print_to_stdout=False):
    printerr(cmd)
    if not args.dry_run:
        returncode, stdout, _ = run(cmd)
        if print_to_stdout:
            printout(stdout)
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
    returncode, stdout, _ = run(cmd, 1)

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
    returncode, stdout, _ = run('minikube status')
    if returncode == 0:
        if ('minikube: Running' in stdout) and ('cluster: Running') in stdout:
            returncode, stdout, _ = run('kubectl config current-context')
            return returncode == 0 and stdout == 'minikube'
    return False


def is_running_minishift():
    returncode, stdout, _ = run('minishift status', show_stderr=False)
    if returncode == 0:
        if ('minishift: Running' in stdout) and ('cluster: Running') in stdout:
            returncode, stdout, _ = run('kubectl config current-context')
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
    returncode, _, _ = run('kubectl version', DEFAULT_TIMEOUT)
    if returncode != 0:
        msg = 'Cannot reach cluster with kubectl: `kubectl version` either timed out or failed to connect (exit code {})'.format(returncode)
        if minishift:
            # Minishift needs special configuration for kubectl to work
            msg = msg + ". Did you do 'eval $(minishift oc-env)'?"
        fail(msg)


def is_int(s):
    try:
        int(s)
        return True
    except:
        return False


def check_credentials(creds):
    registry = 'https://lightbend-docker-commercial-registry.bintray.io/v2'
    api_url = registry + '/enterprise-suite/console-api/tags/list'

    # Use curl for checking credentials by default, only do urllib2 backup in case curl isn't installed
    returncode, stdout, _ = run('curl --version')
    if returncode == 0:
        returncode, stdout, _ = run('curl -s -o /dev/null -w "%{http_code}" ' + ' --user {}:{} {}'
                                    .format(creds[0], creds[1], api_url), DEFAULT_TIMEOUT, show_stderr=True)
        return is_int(stdout) and int(stdout) == 200

    # Set up basic auth with given creds
    req = url.Request(api_url)
    basic_auth = base64.b64encode('{}:{}'.format(creds[0], creds[1]))
    req.add_header('Authorization', 'Basic ' + basic_auth)

    success = False
    try:
        resp = url.urlopen(req)
        if resp.getcode() == 200:
            # Lazy way of verifying returned json - there should be a tag named "latest"
            if '"latest"' in resp.read():
                success = True
    except url.HTTPError as err:
        if err.code != 401:
            printerr('error: check_credentials failed: {}'.format(err))
    except url.URLError as err:
        if err.reason.errno == 54:
            # Code 54 error can be raised when old TLS is used due to old python
            printerr('error: check_credentials TLS authorization failed; this can be due to an old python '
                     'version installed on OS X - please upgrade your python version')
        else:
            printerr('error: check_credentials failed: {}'.format(err))
    finally:
        return success


# compare the contents of current running installer and remote installer.
# Print warning if they are different
def check_new_install_script():
    connect_timeout = 3
    curl_max_tmeout = 5
    installer_url = "https://raw.githubusercontent.com/lightbend/console-charts/master/enterprise-suite/scripts/lbc.py"

    # use curl command as first option.
    returncode, _, _ = run('curl --version')
    if returncode == 0:
        returncode, rmt_installer_cnts, _ = run('curl -s --connect-timeout {} --max-time {} {}'.format(connect_timeout,
                                                                                                       curl_max_tmeout,
                                                                                                       installer_url),
                                                DEFAULT_TIMEOUT, show_stderr=True)
        if returncode != 0:
            return
    else:
        try:
            response = url.urlopen(installer_url, timeout=connect_timeout)
            if response == None:
                return
            rmt_installer_cnts = response.read()
        except url.URLError as e:
            # if we cannot connect to remote server, ignore for now...
            return

    # read the contents of the current installer
    with open(os.path.abspath(__file__)) as f:
        current_installer_contents = f.read()

    if rmt_installer_cnts != current_installer_contents:
        printout("info: New installer is available. Use the following command to download it: curl -O {}".format(installer_url))


def preinstall_check(creds, minikube=False, minishift=False):
    check_helm()
    check_kubectl()
    check_new_install_script()

    if minikube:
        require_version('minikube version', REQ_VER_MINIKUBE)

    if minishift:
        require_version('minishift version', REQ_VER_MINISHIFT)
        require_version('oc version', REQ_VER_OC)

    # Check if helm is set up inside a cluster
    returncode, _, _ = run('helm version', DEFAULT_TIMEOUT)
    if returncode != 0:
        fail('Cannot get helm status. Did you set up helm inside your cluster?')

    # TODO: Check if RBAC rules for tiller are set up

    if not check_credentials(creds):
        printerr('Your credentials might not be correct' +
                 ' - unable to make authenticated request to lightbend docker registry; ' +
                 'proceeding with the installation anyway')


# Returns one of 'deployed', 'failed', 'pending', 'deleting', 'notfound' or 'unknown'
# Also returns the namespace.  Useful for uninstall.
def install_status(release_name):
    returncode, stdout, _ = run('helm status ' + release_name,
                                DEFAULT_TIMEOUT, show_stderr=False)
    namespace = None
    if returncode != 0:
        return 'notfound', namespace

    match = re.search(r'^NAMESPACE: ([\w-]+)', stdout, re.MULTILINE)
    if match:
        namespace = match.group(1)

    if 'STATUS: DEPLOYED' in stdout or (stdout == ''):
        status = 'deployed'
    elif 'STATUS: FAILED' in stdout:
        status = 'failed'
    elif 'STATUS: PENDING_INSTALL' in stdout:
        status = 'pending'
    elif 'STATUS: PENDING_UPGRADE' in stdout:
        status = 'pending'
    elif 'STATUS: DELETING' in stdout:
        status = 'deleting'
    else:
        status = 'unknown'

    return status, namespace


# Helper function that runs a command, then looks for expected strings
# in the output, one per line. Returns True if everything in the 'expected'
# list was found, False if nothing was found. If some resources were found,
# but not all, it fails (sys.exit()) with a given message.
def check_resource_list(cmd, expected, fail_msg):
    returncode, stdout, _ = run(cmd)
    if returncode == 0:
        all_found = True
        found_resources = []
        lines = stdout.split('\n')
        for res in expected:
            found_lines = filter(lambda x: res in x, lines)
            if len(found_lines) == 1:
                found_resources.append(res)
                lines.remove(found_lines[0])
            elif len(found_lines) == 0:
                all_found = False
            else:
                fail('Multiple lines with resource {} found: {}'.format(res, str(found_lines)))

        if not all_found and len(found_resources) > 0:
            fail(fail_msg.format(str(found_resources), str(expected)))

        return all_found
    return False


# Check that the use of persistentVolumes before and after the (un)install will not lead to data loss.
def check_pv_usage(uninstalling=False):
    # Retrieve the current helm installation computed values.
    returncode, helm_get_stdout, helm_get_stderr = run('helm get ' + args.helm_name,
                                                       DEFAULT_TIMEOUT, show_stderr=False)

    if 'usePersistentVolumes' not in values:
        printerr("warning: can't determine the value of usePersistentVolumes, unable to parse helm values")
        return

    wants_pvcs = values['usePersistentVolumes'] is True

    if 'Error: release: "{}" not found'.format(args.helm_name) in helm_get_stderr:
        return

    if returncode != 0:
        printerr("Unable to retrieve release information.  Proceed with caution.")
        return

    has_pvcs = 'usePersistentVolumes: true' in helm_get_stdout

    if has_pvcs:
        if uninstalling:
            printerr("error: Existing deployment has usePersistentVolumes=true. Uninstall may "
                     "result in the loss of Console data as associated PVCs will be removed.")
        elif not wants_pvcs:
            printerr("error: Existing deployment has usePersistentVolumes=true, but upgrade has specified "
                     "usePersistentVolumes=false. Upgrade may result in the loss of Console data as associated "
                     "PVCs will be removed.")
        else:
            # Nothing potentially dangerous here - we are neither uninstalling nor setting usePersistentVolumes=false.
            return

        printerr("error: Invoke with '--delete-pvcs' to proceed")
        fail("error: Stopping")


# Takes helm style "key1=value1,key2=value2" string and returns a list of (key, value)
# pairs. Supports quoting, escaped or non-escaped commas and values with commas inside, eg.:
#   parse_set_string('am=amg01:9093,amg02:9093') -> [('am', 'amg01:9093,amg02:9093')]
#   parse_set_string('am="am01,am02",es=NodePort') -> [('am', 'am01,am02'), ('es', 'NodePort')]
def parse_set_string(s):
    # Keyval pair with commas allowed
    keyval_pair_re = re.compile(r'(\w+)=([\w\-\+\*\:\.\\,]+)')
    # Keyval pair without commas
    keyval_pair_nc_re = re.compile(r'(\w+)=([\w\-\+\*\:\.]+)')
    # Keyval pair with quoted value
    keyval_pair_quot_re = re.compile(r'(\w+)="(.*?)"')

    # We accept either a single keyval pair with commas allowed inside value, or multiple pairs
    m = keyval_pair_re.match(s)
    if m is not None and m.group(0) == s:
        return [(m.group(1), m.group(2).replace('\\,', ','))]
    else:
        left, result = s, []
        while len(left) > 0:
            mq = keyval_pair_quot_re.match(left)
            mn = keyval_pair_nc_re.match(left)
            m = mq or mn
            if m is not None:
                matchlen = len(m.group(0))
                result.append((m.group(1), m.group(2).replace('\\,', ',')))
                if matchlen == len(left) or left[matchlen] == ',':
                    left = left[matchlen + 1:]
                else:
                    raise ValueError('unexpected character "{}"'.format(left[matchlen]))
            else:
                raise ValueError('unable to parse "{}"'.format(left))
        return result


def make_fetchdir():
    """Makes a temporary directory for fetching the helm chart. Used by tests to stub out the directory creation."""
    return tempfile.mkdtemp()


def prune_template_args(helm_args):
    valid_args = r'(?:--set|--values|-f|--set-file|--set-string)(?:\s+|=)[^\s]+'
    res = re.findall(valid_args, helm_args)
    return ' '.join(res)


def install(creds_file):
    creds_arg = '--values ' + creds_file
    version_arg = ''
    if args.version is not None:
        version_arg = ('--version ' + args.version)
    namespace_arg = "--namespace {}".format(args.namespace)

    # Handle --namespace in helm args (after --) to complain about conflicts
    helm_args = ''
    if len(args.rest) > 0:
        hparser = argparse.ArgumentParser()
        hparser.add_argument('--namespace')
        ns_args, unknown = hparser.parse_known_args(args.rest)

        if ns_args.namespace is not None:
            if args.namespace != "lightbend" and args.namespace != ns_args.namespace:
                printerr("warning: conflicting namespace values provided in arguments {} and {} ".format(
                    args.namespace, ns_args.namespace))
                fail("Invoke again with correct namespace value...")
            namespace_arg = ""

        helm_args += ' '.join(args.rest)

    # Add '--set' arguments to helm_args
    if args.set:
        if helm_args:
            helm_args += ' '
        for s in args.set:
            for key, val in parse_set_string(s):
                helm_args += '--set {}={} '.format(key, val.replace(',', '\\,'))

    tempdir = None
    try:
        if args.local_chart:
            chart_name = args.local_chart
            chart_file = args.local_chart
        else:
            chart_name = "{}/{}".format(args.repo_name, args.chart)
            execute('helm repo add {} {}'.format(args.repo_name, args.repo))
            execute('helm repo update')
            tempdir = make_fetchdir()
            chart_file = fetch_remote_chart(tempdir)

        if args.export_yaml:
            # Tillerless path - renders kubernetes resources and prints to stdout.
            creds_exec_arg = ''
            if args.export_yaml == 'creds':
                creds_exec_arg = '--execute templates/commercial-credentials.yaml ' + creds_arg
                printerr('warning: credentials in yaml are not encrypted, only base64 encoded. Handle appropriately.')

            execute('helm template --name {} {} {} {} {}'.format(args.helm_name, namespace_arg,
                                                                 helm_args, creds_exec_arg, chart_file),
                    print_to_stdout=True)

        else:
            # Tiller path - installs console directly to a k8s cluster in a given namespace

            # Calculate computed values for chart to be installed.
            template_args = prune_template_args(helm_args)
            rc, template_stdout, template_stderr = run('helm template -x templates/dump-values.yaml {} {}'.
                                                       format(template_args, chart_file),
                                                       show_stderr=False)
            global values
            if rc != 0:
                printerr("warning: unable to determine computed helm values - this may lead to incorrect warnings")
                values = {}
            else:
                try:
                    computed = template_stdout.splitlines()[-2][2:]
                    values = json.loads(computed)
                except Exception as e:
                    printerr("warning: unable to parse helm values - this may lead to incorrect warnings")
                    printerr(e)
                    values = {}

                printout("Computed chart values:", json.dumps(values, sort_keys=True, indent=2, separators=(',', ': ')))
                printout("")

            # Determine if we should upgrade or install
            should_upgrade = False

            # Check status of existing install under the same release name
            status, namespace = install_status(args.helm_name)
            if status == 'deployed' or status == 'failed':
                if status == 'failed':
                    printerr('info: found a failed installation under name {}, will attempt to upgrade'.format(args.helm_name))
                    printerr('info: if this fails, pass `--force-install` to delete the prior installation first')

                if args.force_install:
                    uninstall(status=status, namespace=namespace)
                    # Give it some time to remove all resources before proceeding with install.
                    # This is not ideal - but neither are custom resource existence checks.
                    # Helm should really be doing this for us.
                    printerr('info: Waiting for prior resources to be removed')
                    time.sleep(REINSTALL_WAIT_SECS)
                else:
                    should_upgrade = True

            elif status == 'notfound':
                # Continue with the install when status is 'notfound'
                pass
            else:
                fail('Unable to install/upgrade console, an install named {} with status {} already exists. '
                     .format(args.helm_name, status))

            if args.wait:
                helm_args += ' --wait'

            if not args.delete_pvcs:
                check_pv_usage()

            if should_upgrade:
                full_args = [args.helm_name, chart_name, version_arg, creds_arg, helm_args]
                full_args = ' '.join(filter(None, full_args))
                execute('helm upgrade {}'.format(full_args))
            else:
                full_args = [args.helm_name, namespace_arg, version_arg, creds_arg, helm_args]
                full_args = ' '.join(filter(None, full_args))
                execute('helm install {} --name {}'.format(chart_name, full_args))
    finally:
        if tempdir:
            shutil.rmtree(tempdir)


def uninstall(status=None, namespace=None):
    if not status:
        status, namespace = install_status(args.helm_name)

    if status == 'notfound':
        fail('Unable to delete console installation - no release named {} found'.format(args.helm_name))
    elif status == 'deleting':
        fail('Unable to delete console installation {} - it is already being deleted'.format(args.helm_name))
    else:
        if not args.delete_pvcs:
            check_pv_usage(uninstalling=True)

        printerr("info: Deleting previous console installation {} with status '{}'".format(args.helm_name, status))
        execute('helm delete --purge ' + args.helm_name)
        printerr('warning: Helm delete does not wait for resources to be fully removed. If a subsequent install fails, '
                 'please re-run it after waiting for all resources to be removed.')


def write_temp_credentials(creds_tempfile, creds):
    creds_str = '\n'.join(["imageCredentials:",
                           "    username: " + creds[0],
                           "    password: " + creds[1]])
    creds_tempfile.write(creds_str)
    creds_tempfile.flush()


def import_credentials():
    creds = (os.environ.get('LIGHTBEND_COMMERCIAL_USERNAME'),
             os.environ.get('LIGHTBEND_COMMERCIAL_PASSWORD'))

    if creds[0] is None or creds[1] is None:
        with open(os.path.expanduser(args.creds), 'r') as creds_file:
            creds_dict = dict(re.findall(r'(\S+)\s*=\s*(".*?"|\S+)', creds_file.read()))
            creds = (creds_dict.get('user'), creds_dict.get('password'))

    if creds[0] is None or creds[1] is None:
        fail("Credentials missing, please check your credentials file\n"
             "LIGHTBEND_COMMERCIAL_CREDENTIALS=" + args.creds)

    return creds


def check_install(external_alertmanager=False):
    def deployment_running(name):
        printout('Checking deployment {} ... '.format(name), end='')
        returncode, stdout, _ = run('kubectl --namespace {} get deploy/{} --no-headers'
                                    .format(args.namespace, name))
        if returncode == 0:
            # Skip first (deployment name) and last (running time) items
            cols = [int(col) for col in stdout.replace('/', ' ').split()[1:-1]]
            desired, _, _, available = cols[0], cols[1], cols[2], cols[3]
            if desired <= 0:
                printout('failed')
                printerr('Deployment {} status check: expected to see 1 or more desired replicas, found 0'
                         .format(name))
            if desired > available:
                printout('failed')
                printerr('Deployment {} status check: available replica number ({}) is less than desired ({})'
                         .format(name, available, desired))
            if desired > 0 and desired <= available:
                printout('ok')
                return True
        else:
            printout('failed')
            printerr('Unable to check deployment {} status'.format(name))
        return False

    def check_deployments(deployments):
        status_ok = True
        deps = deployments
        if not external_alertmanager:
            deps = deps + [CONSOLE_ALERTMANAGER_DEPLOYMENT]

        for dep in deps:
            status_ok &= deployment_running(dep)
        return status_ok

    status_ok = check_deployments(CONSOLE_DEPLOYMENTS) 
    if not status_ok:
        printerr('\nIt appears you might be running older version of console, checking old deployment names...\n')
        status_ok = check_deployments(CONSOLE_DEPLOYMENTS_OLD)

    if status_ok:
        printerr('Your Lightbend Console seems to be running fine!')
    else:
        fail('Lightbend Console status check failed')


def debug_dump(args):
    def dump(dest, filename, content):
        if args.print:
            # Print to stdout
            printout('=== File: {} ==='.format(filename))
            printout(content)
        else:
            # Put to a zipfile
            dest.writestr(filename, content)

    def get_pod_containers(pod):
        # This magic gives us all the containers in a pod
        returncode, containers, _ = run("kubectl --namespace {} get pods {} -o jsonpath='{{.spec.containers[*].name}}'"
                                        .format(args.namespace, pod))
        if returncode == 0:
            return containers.split()
        else:
            fail(failure_msg + 'unable to get containers in a pod {}'
                 .format(pod))

    def write_log(archive, pod, container):
        returncode, logs_out, _ = run('kubectl --namespace {} logs {} -c {} --tail=250'
                                    .format(args.namespace, pod, container),
                                    show_stderr=False)
        if returncode == 0:
            filename = '{}+{}.log'.format(pod, container)
            dump(archive, filename, logs_out)
        else:
            printerr(failure_msg + 'unable to get logs for container {} in a pod {}'
                 .format(container, pod))

        # Try to get previous logs too
        returncode, logs_out, _ = run('kubectl --namespace {} logs {} -c {} -p --tail=250'
                                    .format(args.namespace, pod, container),
                                    show_stderr=False)
        if returncode == 0:
            filename = '{}+{}+prev.log'.format(pod, container)
            dump(archive, filename, logs_out)

    printerr('Capturing debug information, this can take a few minutes')

    failure_msg = 'Failed to get diagnostic data (this is likely harmless): '

    archive = None
    if not args.print:
        timestamp = datetime.datetime.now().strftime('%Y-%m-%d-%H-%M-%S')
        filename = 'console-diagnostics-{}.zip'.format(timestamp)
        archive = zipfile.ZipFile(filename, 'w')

    printerr('Listing Console resources')
    returncode, stdout, _ = run('kubectl --namespace {} get all'.format(args.namespace),
                                show_stderr=False)
    if returncode == 0:
        dump(archive, 'kubectl-get-all.txt', stdout)
    else:
        fail(failure_msg + 'unable to list k8s resources in {} namespace'
             .format(args.namespace))

    printerr('Describing Console resources')
    returncode, stdout, _ = run('kubectl --namespace {} describe all'.format(args.namespace),
                                show_stderr=False)
    if returncode == 0:
        dump(archive, 'kubectl-describe-all.txt', stdout)
    else:
        fail(failure_msg + 'unable to describe k8s resources in {} namespace'
             .format(args.namespace))

    printerr('Describing Console PVCs')
    returncode, stdout, _ = run('kubectl --namespace {} get pvc'.format(args.namespace),
                                show_stderr=False)
    if returncode == 0:
        dump(archive, 'kubectl-get-pvc.txt', stdout)
    else:
        fail(failure_msg + 'unable to describe k8s resources in {} namespace'
             .format(args.namespace))

    printerr('Retrieving Console logs')
    returncode, stdout, _ = run('kubectl --namespace {} get pods --no-headers'.format(args.namespace))
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
        printerr('Please attach this zip file to a support ticket')


def setup_args(argv):
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(dest='subcommand', help='sub-command help')

    fmt = argparse.ArgumentDefaultsHelpFormatter
    install = subparsers.add_parser('install', help='install lightbend console', formatter_class=fmt)
    uninstall = subparsers.add_parser('uninstall', help='uninstall lightbend console', formatter_class=fmt)
    verify = subparsers.add_parser('verify', help='verify console installation', formatter_class=fmt)
    debug_dump = subparsers.add_parser('debug-dump',
                                       help='make an archive with k8s status info for debugging and diagnostic purposes',
                                       formatter_class=fmt)

    # Debug dump arguments
    debug_dump.add_argument('--print', help='print the output instead of writing it to a zip file',
                            action='store_true')

    # Install arguments
    install.add_argument('--force-install',
                         help='set to true to delete an existing install first, instead of upgrading',
                         action='store_true')
    install.add_argument('--export-yaml', help='export resource yaml to stdout',
                         choices=['creds', 'console'])

    install.add_argument('--local-chart', help='set to location of local chart tarball')
    install.add_argument('--chart', help='chart name to install from the repository', default='enterprise-suite')
    install.add_argument('--repo', help='helm chart repository', default='https://repo.lightbend.com/helm-charts')
    install.add_argument('--repo-name', help='name to give helm chart repository when adding to Tiller',
                         default='es-repo')
    install.add_argument('--creds', help='credentials file', default=os.path.join('~', '.lightbend',
                                                                                  'commercial.credentials'))
    install.add_argument('--version', help='console version to install', type=str)
    install.add_argument('--wait', help='wait for install to finish before returning',
                         action='store_true')
    install.add_argument('--set', help='set a helm chart value, can be repeated for multiple values', type=str,
                         action='append')

    # Verify arguments
    verify.add_argument('--external-alertmanager',
                        help='skips alertmanager check (for use with existing alertmanagers)',
                        action='store_true')

    # Common arguments for install and uninstall
    for subparser in [install, uninstall]:
        subparser.add_argument('--delete-pvcs', help='ignore warnings about PVs and proceed anyway. CAUTION!',
                               action='store_true')
        subparser.add_argument('--dry-run', help='only print out the commands that will be executed',
                               action='store_true')
        subparser.add_argument('--helm-name', help='helm release name', default='enterprise-suite')
        subparser.add_argument('rest',
                               help="any additional arguments separated by '--' will be passed to helm (eg. '-- --set usePersistentVolumes=false')",
                               nargs='*')

    # Common arguments for install, verify and dump
    for subparser in [install, verify, debug_dump]:
        subparser.add_argument('--namespace', help='namespace to install console into/where it is installed',
                               required=True)

    # Common arguments for all subparsers
    for subparser in [install, uninstall, verify, debug_dump]:
        subparser.add_argument('--skip-checks', help='skip environment checks',
                               action='store_true')

    try:
        args = parser.parse_args(argv)
    except:
        fail("")

    if len(argv) == 0:
        parser.print_help()

    return args


def fetch_remote_chart(destdir):
    """" Fetch remote chart into destdir and return its file location. """
    extra_args = ""
    if args.version:
        extra_args = "--version %s" % args.version
    chart_url = 'helm fetch --destination {} {} {} {}/{}'\
        .format(destdir, extra_args, args.repo, args.repo_name, args.chart)
    rc, fetch_stdout, fetch_stderr = run(chart_url, DEFAULT_TIMEOUT, show_stderr=False)
    if rc != 0:
        printerr("unable to reach helm repo: ", chart_url)
        printerr(fetch_stderr)
        fail("")
    chart = glob.glob(destdir + '/enterprise-suite-*.tgz')[0]
    return chart


def main(argv):
    global args
    args = setup_args(argv)

    force_verify = False
    if args.subcommand == 'install':
        creds = import_credentials()

        if not args.skip_checks:
            if args.export_yaml == None:
                minikube = is_running_minikube()
                minishift = is_running_minishift()
                preinstall_check(creds, minikube, minishift)
            else:
                check_helm()

        if args.version == None and args.local_chart == None:
            printerr(("warning: --version has not been set, helm will use the latest available version. "
                      "It is recommended to use an explicit version."))

        # Cannot use `with` statement to auto-delete here because on windows we cannot have file opened by two processes at the same time
        creds_tempfile = tempfile.NamedTemporaryFile('w', delete=False)
        filename = creds_tempfile.name
        if windows:
            # Escape path separators on windows because helm does its own un-escaping
            filename = filename.replace('\\', '\\\\')
        try:
            write_temp_credentials(creds_tempfile, creds)
            creds_tempfile.close()
            install(filename)
        finally:
            os.remove(filename)

        if args.wait:
            force_verify = True

    if args.subcommand == 'verify' or force_verify:
        if not args.skip_checks:
            check_kubectl()
        if force_verify:
            check_install()
        else:
            check_install(args.external_alertmanager)

    if args.subcommand == 'uninstall':
        if not args.skip_checks:
            check_helm()
        uninstall()

    if args.subcommand == 'debug-dump':
        if not args.skip_checks:
            check_kubectl()
        debug_dump(args)


if __name__ == '__main__':
    main(sys.argv[1:])
