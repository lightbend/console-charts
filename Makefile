# Top-level Makefile for creating the index of helm charts.

# ALL CHARTS MUST:
# - Have a Makefile with a recipe for a 'test' target.  This target is run by Travis
#   to test the release.
#   That could be as simple as (the non-testing) Makefile:
#      test: ;
# - Be responsible for putting their chart package files in the docs directory

# Collection of charts to process.
# These are subdirectories of helm-charts.  Add to the list as required.
#
# This setting is the default.  To build a subset (at this level), just define a
# value on the command line.  e.g. "make CHARTS=sample-project"
#
CHARTS = enterprise-suite enterprise-suite-latest reactive-sandbox #sample-project


docs/index.yaml: init $(wildcard docs/*.tgz)
	helm repo index docs --url https://lightbend.github.io/helm-charts

init:
	@helm init -c > /dev/null

test: $(CHARTS)

# All charts must implement the test target
$(CHARTS):
	$(info *** making $(MAKECMDGOALS) on $@)
	$(MAKE) -C $@ $(MAKECMDGOALS)

clean:
	rm docs/index.yaml

.PHONY: $(CHARTS) init test clean
