#!/usr/bin/env python

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
    if re.match(cmd_re, cmd) != None:
        return returncode, stdout, None
    raise UnexpectedCmdException("Expected a command that matches '{}', instead got '{}'.\n\nTest traceback:\n{}".format(cmd_re, cmd, tb))


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


class HelmCommandsTest(unittest.TestCase):
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
            creds.write('realm = Bintray\nhost = dl.bintray.com\nuser = aaa\npassword = bbb\n')

        self.setUpFakeChartfile()

        expected_cmds.clear()

    def tearDown(self):
        # Remove credentials tempdir
        shutil.rmtree(self.creds_dir)

        self.assertEqual(len(expected_cmds), 0)

    def test_export_yaml_console(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend\s+\S+\.tgz')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--export-yaml=console'])

    def test_export_yaml_local_chart(self):
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend   chart\.tgz')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--export-yaml=console', '--local-chart=chart.tgz'])

    def test_export_yaml_creds(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend  --execute templates/commercial-credentials\.yaml\s+--values \S+ \S+\.tgz')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--export-yaml=creds'])

    def test_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend\s+--values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])
    
    def test_install_wait(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend\s+--values \S+\s+--wait')
        
        # Verify happens automatically when --wait is provided
        expect_cmd(r'kubectl --namespace lightbend get deploy/console-backend --no-headers',
                   stdout='console-backend 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/console-frontend --no-headers',
                   stdout='console-frontend 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/grafana --no-headers',
                   stdout='grafana 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--wait'])

    def test_install_helm_failed(self):
        # Failed previous install, no PVCs or clusterroles found for reuse
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: FAILED\nNOTES: blah')
        expect_cmd(r'helm upgrade enterprise-suite es-repo/enterprise-suite\s+--values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])

    def test_install_not_finished(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: PENDING_INSTALL\nNOTES: blah')

        # Expect install to fail when previous install is still pending
        with self.assertRaises(TestFailException):
            lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file])

    def test_upgrade(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm upgrade enterprise-suite es-repo/enterprise-suite\s+--values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file])

    def test_force_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm delete --purge enterprise-suite')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend\s+--values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--force-install', '--delete-pvcs'])

    def test_helm_args(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend\s+--values \S+ --set minikube=true --fakearg')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg'])

    def test_helm_args_namespace(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --values \S+ --set minikube=true --fakearg --namespace=foobar')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg', '--namespace=foobar'])

    def test_helm_args_namespace_val(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --values \S+ --set minikube=true --fakearg --namespace foobar')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--', '--set', 'minikube=true', '--fakearg', '--namespace', 'foobar'])

    def test_helm_args_conflicting_namespace(self):
        with self.assertRaises(TestFailException):
            lbc.main(['install', '--namespace=lightbend', '--namespace', 'foo', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs',
                      '--', '--set', 'minikube=true', '--fakearg', '--namespace', 'bar'])
    
    # At some point lbc ate `--timeout` and passed to helm only the number 110
    def test_helm_args_timeout(self):
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install . --name enterprise-suite --namespace foo --values \S+ --timeout 110')
        lbc.main(['install', '--namespace=lightbend', '--local-chart', '.', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--namespace', 'foo', '--', '--timeout', '110'])


    def test_helm_set(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --values \S+ --set minikube=true --set usePersistentVolumes=true ')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--set', 'minikube=true', '--set', 'usePersistentVolumes=true'])

    def test_helm_set_array(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --values \S+ --set alertmanagers=alertmgr-00\\,alertmgr-01\\,alertmgr-02 ')

        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--creds='+self.creds_file, '--delete-pvcs', '--set', 'alertmanagers=alertmgr-00,alertmgr-01,alertmgr-02'])

    def test_specify_version(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --version 1\.0\.0-rc\.9 --values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--version=1.0.0-rc.9'])

    def test_install_local_chart(self):
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install chart.tgz --name enterprise-suite --namespace lightbend --values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--local-chart=chart.tgz'])

    def test_install_override_repo(self):
        expect_cmd(r'helm repo add es-repo https://repo.bintray.com/helm')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.bintray.com/helm es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--repo=https://repo.bintray.com/helm'])

    def test_install_override_name(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch .* https://repo.lightbend.com/helm-charts es-repo/enterprise-suite')
        expect_cmd(r'helm template .*')
        expect_cmd(r'helm status lb-console', returncode=-1)
        expect_cmd(r'helm install es-repo/enterprise-suite --name lb-console --namespace lightbend --values \S+')
        lbc.main(['install', '--namespace=lightbend', '--skip-checks', '--delete-pvcs', '--creds='+self.creds_file, '--helm-name=lb-console'])

    def test_uninstall(self):
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                  stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: DEPLOYED\nNOTES: blah')
        expect_cmd(r'helm delete --purge enterprise-suite')
        lbc.main(['uninstall', '--skip-checks', '--delete-pvcs'])

    def test_uninstall_not_found(self):
        expect_cmd(r'helm status enterprise-suite', returncode=-1)

        # Expect uninstall to fail
        with self.assertRaises(TestFailException):
            lbc.main(['uninstall', '--skip-checks'])

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
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

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
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

        lbc.main(['verify', '--skip-checks', '--namespace=monitoring'])


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
