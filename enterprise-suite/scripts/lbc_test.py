#!/usr/bin/env python

import lbc
import re
import collections
import tempfile
import shutil
import unittest

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
    expected_cmds.append((re_pattern, returncode, stdout))

def test_run(cmd, timeout=None, stdin=None, show_stderr=True):
    cmd_re, returncode, stdout = expected_cmds.popleft()
    if re.match(cmd_re, cmd) != None:
        return stdout, returncode
    raise UnexpectedCmdException("Expected a command that matches '{}', instead got '{}'"
                                 .format(cmd_re, cmd))
                                
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

class SetStrParsingTest(unittest.TestCase):
    def test_parsing(self):
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


class LbcTest(unittest.TestCase):
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

        expected_cmds.clear()

    def tearDown(self):
        # Remove credentials tempdir
        shutil.rmtree(self.creds_dir)

        self.assertEqual(len(expected_cmds), 0)

    def test_export_yaml_console(self):
        self.setUpFakeChartfile()
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch -d \S* --devel es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend   \S*\.tgz')
        lbc.main(['install', '--skip-checks', '--export-yaml=console'])

    def test_export_yaml_local_chart(self):
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend   chart\.tgz')
        lbc.main(['install', '--skip-checks', '--export-yaml=console', '--local-chart=chart.tgz'])

    def test_export_yaml_creds(self):
        self.setUpFakeChartfile()
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm fetch -d \S* --devel es-repo/enterprise-suite')
        expect_cmd(r'helm template --name enterprise-suite --namespace lightbend  --execute templates/commercial-credentials\.yaml --values \S+ \S+\.tgz')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--export-yaml=creds'])

    def test_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])
    
    def test_install_wait(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+  --wait')
        
        # Verify happens automatically when --wait is provided
        expect_cmd(r'kubectl --namespace lightbend get deploy/es-console --no-headers',
                   stdout='es-console 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/grafana-server --no-headers',
                   stdout='grafana-server 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-server --no-headers',
                   stdout='prometheus-server 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace lightbend get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--wait'])

    def test_install_helm_failed(self):
        # Failed previous install, no PVCs or clusterroles found for reuse
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: FAILED\nNOTES: blah')
        expect_cmd(r'helm delete --purge enterprise-suite')
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])

    def test_install_helm_failed_reuse(self):
        # Failed previous install, PVCs found for reuse
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: FAILED\nNOTES: blah')
        expect_cmd(r'helm delete --purge enterprise-suite')
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers',
                   stdout=('alertmanager-storage Bound pvc-a3815792-e744-11e8-a15b-080027dccb43 32Gi RWO standard 1h\n'
                           'es-grafana-storage Bound pvc-a3824cbc-e744-11e8-a15b-080027dccb43 32Gi RWO standard 1h\n'
                           'prometheus-storage Bound pvc-a382f4c1-e744-11e8-a15b-080027dccb43 256Gi RWO standard 1h\n'))
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])

    def test_install_not_finished(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                   stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: PENDING_INSTALL\nNOTES: blah')

        # Expect install to fail when previous install is still pending
        with self.assertRaises(TestFailException):
            lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])

    def test_upgrade(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm upgrade enterprise-suite es-repo/enterprise-suite --devel --values \S+ ')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file])

    def test_force_install(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=0)
        expect_cmd(r'helm delete --purge enterprise-suite')
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--force-install'])

    def test_helm_args(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+ --set minikube=true --fakearg')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--', '--set', 'minikube=true', '--fakearg'])

    def test_helm_set(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+ --set minikube=true --set usePersistentVolumes=true ')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--set', 'minikube=true', '--set', 'usePersistentVolumes=true'])

    def test_helm_set_array(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+ --set alertmanagers=alertmgr-00\\,alertmgr-01\\,alertmgr-02 ')

        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--set', 'alertmanagers=alertmgr-00,alertmgr-01,alertmgr-02'])

    def test_specify_version(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --version 1\.0\.0-rc\.9 --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--version=1.0.0-rc.9'])

    def test_install_local_chart(self):
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install chart.tgz --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--local-chart=chart.tgz'])

    def test_install_override_repo(self):
        expect_cmd(r'helm repo add es-repo https://repo.bintray.com/helm')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status enterprise-suite', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name enterprise-suite --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--repo=https://repo.bintray.com/helm'])

    def test_install_override_name(self):
        expect_cmd(r'helm repo add es-repo https://repo.lightbend.com/helm-charts')
        expect_cmd(r'helm repo update')
        expect_cmd(r'helm status lb-console', returncode=-1)
        expect_cmd(r'kubectl get pvc --namespace=lightbend --no-headers')
        expect_cmd(r'helm install es-repo/enterprise-suite --name lb-console --namespace lightbend --devel --values \S+')
        lbc.main(['install', '--skip-checks', '--creds='+self.creds_file, '--helm-name=lb-console'])

    def test_uninstall(self):
        expect_cmd(r'helm status enterprise-suite', returncode=0,
                  stdout='LAST DEPLOYED: Tue Nov 13 09:59:46 2018\nNAMESPACE: lightbend\nSTATUS: DEPLOYED\nNOTES: blah')
        expect_cmd(r'helm delete --purge enterprise-suite')
        lbc.main(['uninstall', '--skip-checks'])

    def test_uninstall_not_found(self):
        expect_cmd(r'helm status enterprise-suite', returncode=-1)

        # Expect uninstall to fail
        with self.assertRaises(TestFailException):
            lbc.main(['uninstall', '--skip-checks'])

    def test_verify_fail(self):
        expect_cmd(r'kubectl --namespace monitoring get deploy/es-console --no-headers',
                   stdout='es-console 1 1 1 1 15m')
        # Grafana is not running
        expect_cmd(r'kubectl --namespace monitoring get deploy/grafana-server --no-headers',
                   stdout='grafana-server 1 1 1 0 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-server --no-headers',
                   stdout='prometheus-server 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

        # Expect verify to fail
        with self.assertRaises(TestFailException):
            lbc.main(['verify', '--skip-checks', '--namespace=monitoring'])

    def test_verify_success(self):
        expect_cmd(r'kubectl --namespace monitoring get deploy/es-console --no-headers',
                   stdout='es-console 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/grafana-server --no-headers',
                   stdout='grafana-server 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-kube-state-metrics --no-headers',
                   stdout='prometheus-kube-state-metrics 1 1 1 1 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-server --no-headers',
                   stdout='prometheus-server 2 2 2 2 15m')
        expect_cmd(r'kubectl --namespace monitoring get deploy/prometheus-alertmanager --no-headers',
                   stdout='prometheus-alertmanager 1 1 1 1 15m')

        lbc.main(['verify', '--skip-checks', '--namespace=monitoring'])

if __name__ == '__main__':
    # Override internal functions so we can test them
    lbc.run = test_run
    lbc.fail = test_fail
    lbc.printerr = test_print
    lbc.printinfo = test_print
    lbc.make_tempdir = test_make_tempdir

    # Run tests
    unittest.main()
