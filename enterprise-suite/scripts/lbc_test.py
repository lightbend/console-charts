#!/usr/bin/env python

from abc import ABCMeta, abstractmethod, abstractproperty
import lbc
import re
import collections
import tempfile
import shutil
import unittest
import traceback


# Exception that gets thrown when unexpected command is
# executed by lbc script.
class UnexpectedCmdException(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)


# Exception that will get thrown in place of sys.exit()
class TestFailException(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)


def test_fail(msg):
    raise TestFailException(msg)


# A FIFO queue of expected commands will be kept here
expected_cmds = collections.deque()


# Adds a command to the list
def expect_cmd(re_pattern, returncode=0, stdout=''):
    # find line number in test for later reporting on a failure
    stack = traceback.format_stack()[-2]
    expected_cmds.append((re_pattern, returncode, stdout, stack))


def test_run(cmd, timeout=None, stdin=None, show_stderr=True):
    cmd_re, returncode, stdout, tb = expected_cmds.popleft()
    cmd_re = '^' + cmd_re + '$'
    if re.match(cmd_re, cmd) != None:
        return returncode, stdout, None
    raise UnexpectedCmdException("Expected a command that matches:\n'{}'\ninstead got:\n'{}'\n\nTest traceback:\n{}".format(cmd_re, cmd, tb))


# Checks if all commands were executed, clears queue
# so that further tests do not get affected by a failing test.
def finish_test(testcase):
    testcase.assertEqual(len(expected_cmds), 0)
    expected_cmds.clear()


def test_print(*args, **kwargs):
    pass


tempdir = ''


def test_make_tempdir():
    return tempdir


class LbcTest(unittest.TestCase):
    def test_parse_set_string(self):
        # Empty
        self.assertEqual(lbc.parse_set_string(""), [])

        # Most trivial case
        self.assertEqual(lbc.parse_set_string("key=value"), [("key", "value")])

        # Escaped commas
        self.assertEqual(lbc.parse_set_string("key=val1\\,val2"), [("key", "val1,val2")])

        # Multiple key value pairs
        self.assertEqual(lbc.parse_set_string("key1=val1,key2=val2"), [("key1", "val1"), ("key2", "val2")])
        self.assertEqual(lbc.parse_set_string("key1=val1,key2=val2,b=c"), [("key1", "val1"), ("key2", "val2"), ("b", "c")])

        # Quoted values
        self.assertEqual(lbc.parse_set_string("key=\"value\""), [("key", "value")])
        self.assertEqual(lbc.parse_set_string("key=\"val1,val2\""), [("key", "val1,val2")])

        # Quoted values + multiple pairs
        self.assertEqual(lbc.parse_set_string("key1=\"val1,val2\",key2=val3"), [("key1", "val1,val2"), ("key2", "val3")])

        # Error cases
        with self.assertRaises(ValueError):
            lbc.parse_set_string("key=val1,val2,key2=val3")

        with self.assertRaises(ValueError):
            lbc.parse_set_string("val1,val2")

        with self.assertRaises(ValueError):
            lbc.parse_set_string("key=val1,key2=val1,val2")

        with self.assertRaises(ValueError):
            lbc.parse_set_string("key=val1,key2==val1")

    def test_prune_template_args(self):
        expected = "--set a=1,b=2 --set=c=3 --values a.yml,b.yml --values=c.yml -f d.yml --set-file e=e.yml,f=f.yml " \
                   "--set-file=g=g.yml --set-string h=4 --set-string=h=4"
        helm_args = "--unsupported " + expected + " --wait"

        self.assertEqual(expected, lbc.prune_template_args(helm_args))


class HelmCommandsTest():
    __metaclass__ = ABCMeta

    def setUpFakeChartfile(self):
        # Create tempdir & fake chartfile there for export yaml tests.
        # They will be removed by lbc itself, so no need to clean up.
        global tempdir
        tempdir = tempfile.mkdtemp()
        open(tempdir + '/enterprise-suite-1.0.0-rc.4.tgz', 'w').close()

    def setUp(self):
        # Create tempdir & credentials file
        self.creds_dir = tempfile.mkdtemp()
        self.creds_file = self.creds_dir + '/commercial.credentials'
        with open(self.creds_file, 'w') as creds:
            creds.write('realm = Cloudsmith API\nhost = commercial-registry.lightbend.com\nuser = aaa\npassword = bbb\n')

        self.setUpFakeChartfile()

        expected_cmds.clear()

    def tearDown(self):
        # Remove credentials tempdir
        shutil.rmtree(self.creds_dir)

        self.assertEqual(len(expected_cmds), 0)

    @abstractproperty
    def helm_version(self):
        pass

    def expect_helm_version(self):
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)

    @abstractmethod
    def expect_helm_status_notfound(self, name='enterprise-suite', namespace='lightbend'):
        pass

    @abstractmethod
    def expect_helm_status_status(self, status):
        pass

    @abstractmethod
    def expect_helm_status_deployed(self):
        pass

    @abstractmethod
    def expect_helm_status_failed(self):
        pass

    @abstractmethod
    def expect_helm_status_pending_install(self):
        pass

    @abstractmethod
    def expect_helm_template(self, chart=r'\S+\.tgz', template=None, args=[]):
        pass

    @abstractmethod
    def expect_helm_install(self, name='enterprise-suite', chart='es-repo/enterprise-suite', namespace='lightbend', version=None, args=[]):
        pass

    @abstractmethod
    def expect_helm_upgrade(self):
        pass

    @abstractmethod
    def expect_helm_delete(self):
        pass

    def expect_basic_setup_checks(self):
        expect_cmd('minikube status')
        expect_cmd('minishift status')
        self.expect_helm_version()
        expect_cmd('kubectl version --client=true --short=true')
        expect_cmd('kubectl version')
        expect_cmd('curl --version')
        expect_cmd(r'curl .*')

    def expect_install_setup(self, repo='https://repo.lightbend.com/helm-charts', setup_checks=None):
        self.expect_helm_version()
        if setup_checks:
            setup_checks()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd('helm repo add es-repo ' + repo)
        expect_cmd('helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')

    @abstractmethod
    def uninstall(self, args=[]):
        pass

    def test_export_yaml_console(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        self.expect_helm_template()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--export-yaml=console', '--creds='+self.creds_file])

    def test_export_yaml_local_chart(self):
        self.expect_helm_version()
        self.expect_helm_template(chart='chart.tgz')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--export-yaml=console', '--local-chart=chart.tgz'], '--creds='+self.creds_file])

    def test_export_yaml_creds(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        self.expect_helm_template(template=r'templates/commercial-credentials\.yaml', args=['--values', r'\S+'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--export-yaml=creds'])

    def test_install(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])

    def test_install_wait(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--wait'])

        # Verify happens automatically when --wait is provided
        expect_cmd(r'kubectl --namespace lightbend get deploy/console-backend --no-headers',
                   stdout='console-backend 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/console-frontend --no-headers',
                   stdout='console-frontend 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/grafana --no-headers',
                   stdout='grafana 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')

        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--wait'])

    def test_install_helm_failed(self):
        # Failed previous install, no PVCs or clusterroles found for reuse
        self.expect_install_setup()
        self.expect_helm_status_failed()
        self.expect_helm_upgrade()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])

    def test_install_not_finished(self):
        self.expect_install_setup()
        self.expect_helm_status_pending_install()

        # Expect install to fail when previous install is still pending
        with self.assertRaises(TestFailException):
            lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file])

    def test_upgrade(self):
        self.expect_install_setup()
        self.expect_helm_status_deployed()
        self.expect_helm_upgrade()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])

    def test_force_install(self):
        self.expect_install_setup()
        self.expect_helm_status_deployed()
        self.expect_helm_delete()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--force-install', '--delete-pvcs'])

    def test_helm_args(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--set', 'minikube=true', '--fakearg'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg'])

    def test_helm_args_namespace(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--set', 'minikube=true', '--fakearg', '--namespace=foobar'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg', '--namespace=foobar'])

    def test_helm_args_namespace_val(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--set', 'minikube=true', '--fakearg', '--namespace', 'foobar'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg', '--namespace', 'foobar'])

    def test_helm_args_conflicting_namespace(self):
        with self.assertRaises(TestFailException):
            self.expect_helm_version()
            lbc.main(['install', '--namespace=lightbend', '--namespace', 'foo', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs',
                      '--', '--set', 'minikube=true', '--fakearg', '--namespace', 'bar'])

    # At some point lbc ate `--timeout` and passed to helm only the number 110
    def test_helm_args_timeout(self):
        self.expect_helm_version()
        expect_cmd(r'helm template .*')
        self.expect_helm_status_notfound(namespace='foo')
        self.expect_helm_install(chart='.', namespace='foo', args=['--timeout', '110'])
        lbc.main(['install', '--namespace=lightbend', '--local-chart', '.', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--namespace', 'foo', '--', '--timeout', '110'])

    def test_helm_set(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--set', 'minikube=true', '--set', 'usePersistentVolumes=true'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--set', 'minikube=true', '--set', 'usePersistentVolumes=true'])

    def test_helm_set_array(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(args=['--set', r'alertmanagers=alertmgr-00\\,alertmgr-01\\,alertmgr-02'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--set', 'alertmanagers=alertmgr-00,alertmgr-01,alertmgr-02'])

    def test_specify_version(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound()
        self.expect_helm_install(version=r'1\.0\.0-rc\.9')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--version=1.0.0-rc.9'])

    def test_install_local_chart(self):
        self.expect_helm_version()
        expect_cmd(r'helm template .*')
        self.expect_helm_status_notfound()
        self.expect_helm_install(chart='chart.tgz')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--local-chart=chart.tgz'])

    def test_install_override_repo(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        self.expect_helm_status_notfound()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--repo=https://repo.lightbend.com/helm'])

    def test_install_override_name(self):
        self.expect_install_setup()
        self.expect_helm_status_notfound('lb-console')
        self.expect_helm_install(name='lb-console')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--helm-name=lb-console'])

    def test_uninstall(self):
        self.expect_helm_version()
        self.expect_helm_status_deployed()
        self.expect_helm_delete()
        self.uninstall(['--skip-checks', '--delete-pvcs'])

    def test_uninstall_not_found(self):
        self.expect_helm_version()
        self.expect_helm_status_notfound()

        # Expect uninstall to fail
        with self.assertRaises(TestFailException):
            self.uninstall(['--skip-checks'])

    def test_verify_fail(self):
        expect_cmd(r'kubectl --namespace monitoring get deploy/console-backend --no-headers',
                   stdout='console-backend 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/console-frontend --no-headers',
                   stdout='console-frontend 1 1 1 1 15m')
        # Grafana is not running
        expect_cmd(r'kubectl --namespace monitoring get deploy/grafana --no-headers',
                   stdout='grafana 1 1 1 0 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')

        # Installer will check old deployment names if the new ones fail
        expect_cmd(r'kubectl --namespace monitoring get deploy/es-console --no-headers',
                   stdout='console-backend 2 2 2 2 15m')
        # Grafana is not running
        expect_cmd(r'kubectl --namespace monitoring get deploy/grafana-server --no-headers',
                   stdout='console-frontend 1 1 1 0 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-server --no-headers',
                   stdout='grafana 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')

        # Expect verify to fail
        with self.assertRaises(TestFailException):
            lbc.main(['verify', '--skip-checks', '--namespace=monitoring'])

    def test_verify_success(self):
        expect_cmd(r'kubectl --namespace monitoring get deploy/console-backend --no-headers',
                   stdout='console-backend 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/console-frontend --no-headers',
                   stdout='console-frontend 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/grafana --no-headers',
                   stdout='grafana 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')

        lbc.main(['verify', '--skip-checks', '--namespace=monitoring'])


class HelmV2CommandsTest(HelmCommandsTest, unittest.TestCase):
    @property
    def helm_version(self):
        return 'Client: v2.16.1+gbbdfe5e'

    def expect_helm_status_notfound(self, name='enterprise-suite', namespace='lightbend'):
        expect_cmd('helm status ' + name, returncode=-1)

    def expect_helm_status_status(self, status):
        expect_cmd('helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: {}\nNOTES: blah'.format(status))

    def expect_helm_status_deployed(self):
        self.expect_helm_status_status('DEPLOYED')

    def expect_helm_status_failed(self):
        self.expect_helm_status_status('FAILED')

    def expect_helm_status_pending_install(self):
        self.expect_helm_status_status('PENDING_INSTALL')

    def expect_helm_template(self, chart=r'\S+\.tgz', template=None, args=[]):
        full_args = []
        if template:
            full_args.extend(['--execute', template])
        full_args.extend(args)
        full_args.append(chart)
        full_args = ' '.join(filter(None, full_args))
        expect_cmd('helm template --name enterprise-suite --namespace lightbend ' + full_args)

    def expect_helm_install(self, name='enterprise-suite', chart='es-repo/enterprise-suite', namespace='lightbend', version=None, args=[]):
        full_args = [chart, '--name', name]
        if not any(item.startswith('--namespace') for item in args):
            full_args.extend(['--namespace', namespace])
        if version:
            full_args.extend(['--version', version])
        full_args.extend(['--values', r'\S+'])
        full_args.extend(args)
        full_args = ' '.join(filter(None, full_args))
        expect_cmd('helm install ' + full_args)

    def expect_helm_upgrade(self):
        expect_cmd(r'helm upgrade enterprise-suite es-repo/enterprise-suite --values \S+')

    def expect_helm_delete(self):
        expect_cmd('helm delete --purge enterprise-suite')

    def uninstall(self, args=[]):
        lbc.main(['uninstall'] + args)

    def test_not_checks_detect_legacy_installs(self):
        def setup_checks():
            self.expect_basic_setup_checks()
            expect_cmd('curl --version')
            expect_cmd(r'curl .*')
        self.expect_install_setup(setup_checks=setup_checks)
        self.expect_helm_status_notfound()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--delete-pvcs', '--creds='+self.creds_file])



class HelmV3CommandsTest(HelmCommandsTest, unittest.TestCase):
    @property
    def helm_version(self):
        return 'v3.0.0+ge29ce2a'

    def expect_helm_status_notfound(self, name='enterprise-suite', namespace='lightbend'):
        expect_cmd('helm status --namespace {} {}'.format(namespace, name), returncode=-1)

    def expect_helm_status_status(self, status):
        expect_cmd('helm status --namespace lightbend enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: {}\nNOTES: blah'.format(status))

    def expect_helm_status_deployed(self):
        self.expect_helm_status_status('deployed')

    def expect_helm_status_failed(self):
        self.expect_helm_status_status('failed')

    def expect_helm_status_pending_install(self):
        self.expect_helm_status_status('pending_install')

    def expect_helm_template(self, chart=r'\S+\.tgz', template=None, args=[]):
        full_args = []
        if template:
            full_args.extend(['--show-only', template])
        full_args.extend(args)
        full_args.append(chart)
        full_args = ' '.join(filter(None, full_args))
        expect_cmd('helm template enterprise-suite --namespace lightbend ' + full_args)

    def expect_helm_install(self, name='enterprise-suite', chart='es-repo/enterprise-suite', namespace='lightbend', version=None, args=[]):
        full_args = [name, chart]
        if not any(item.startswith('--namespace') for item in args):
            full_args.extend(['--namespace', namespace])
        if version:
            full_args.extend(['--version', version])
        full_args.extend(['--values', r'\S+'])
        full_args.extend(args)
        full_args = ' '.join(filter(None, full_args))
        expect_cmd('helm install ' + full_args)

    def expect_helm_upgrade(self):
        expect_cmd(r'helm upgrade enterprise-suite --namespace lightbend es-repo/enterprise-suite --values \S+')

    def expect_helm_delete(self):
        expect_cmd('helm delete --namespace lightbend enterprise-suite')

    def uninstall(self, args=[]):
        lbc.main(['uninstall', '--namespace', 'lightbend'] + args)

    def test_checks_detect_legacy_installs(self):
        def setup_checks():
            self.expect_basic_setup_checks()
            expect_cmd('kubectl get configmap --namespace kube-system --selector OWNER=TILLER,NAME=enterprise-suite --ignore-not-found=true')
            expect_cmd('curl --version')
            expect_cmd(r'curl .*')
        self.expect_install_setup(setup_checks=setup_checks)
        self.expect_helm_status_notfound()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--delete-pvcs', '--creds='+self.creds_file])

    def test_checks_detect_legacy_installs_exist(self):
        with self.assertRaises(TestFailException):
            self.expect_helm_version()
            self.expect_basic_setup_checks()
            expect_cmd('kubectl get configmap --namespace kube-system --selector OWNER=TILLER,NAME=enterprise-suite --ignore-not-found=true',
                       stdout='NAME                  DATA   AGE\nenterprise-suite.v1   1      7s')
            lbc.main(['install', '--namespace=lightbend', '--delete-pvcs', '--creds='+self.creds_file])

    def test_checks_detect_legacy_installs_fail(self):
        with self.assertRaises(TestFailException):
            self.expect_helm_version()
            self.expect_basic_setup_checks()
            expect_cmd('kubectl get configmap --namespace kube-system --selector OWNER=TILLER,NAME=enterprise-suite --ignore-not-found=true',
                       returncode=1)
            lbc.main(['install', '--namespace=lightbend', '--delete-pvcs', '--creds='+self.creds_file])

class HelmV332AndAboveCommandsTest(HelmV3CommandsTest, unittest.TestCase):
    @property
    def helm_version(self):
        return 'v3.3.2+ga61ce56'

    def expect_install_setup(self, repo='https://repo.lightbend.com/helm-charts', setup_checks=None):
        self.expect_helm_version()
        if setup_checks:
            setup_checks()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        # For Helm above 3.3.2, it is necessary to use --force-update
        expect_cmd('helm repo add --force-update es-repo ' + repo)
        expect_cmd('helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')


    def expect_helm_status_status(self, status):
        expect_cmd('helm status --namespace lightbend enterprise-suite', returncode=0,
                   stdout='NAME: enterprise-suite\nLAST DEPLOYED: Mon Oct  5 16:41:58 2020\nNAMESPACE: lightbend\nSTATUS: {}\nREVISION: 1\nTEST SUITE: None\nNOTES: Thank you for installing enterprise-suite.'.format(status))

    def test_install_override_repo(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add --force-update es-repo https://repo.lightbend.com/helm')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        self.expect_helm_status_notfound()
        self.expect_helm_install()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--repo=https://repo.lightbend.com/helm'])

    def test_export_yaml_creds(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add --force-update es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        self.expect_helm_template(template=r'templates/commercial-credentials\.yaml', args=['--values', r'\S+'])
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--export-yaml=creds'])

    def test_export_yaml_console(self):
        self.expect_helm_version()
        expect_cmd(r'helm version --client --short', stdout=self.helm_version)
        expect_cmd(r'helm repo add --force-update es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* es-repo/enterprise-suite')
        self.expect_helm_template()
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--export-yaml=console'], '--creds='+self.creds_file])


if __name__ == '__main__':
    # Override internal functions so we can test them
    lbc.run = test_run
    lbc.fail = test_fail
    lbc.printerr = test_print
    lbc.printout = test_print
    lbc.make_fetchdir = test_make_tempdir
    lbc.REINSTALL_WAIT_SECS = 0

    # Run tests
    unittest.main()
