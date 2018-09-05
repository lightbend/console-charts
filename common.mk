##
# Chart Makefiles must implement package and test targets.  They are free to implement
# those however they want.  If desired, the Makefile can include this file that provides
# default recipes for those targets
##

HELM_CHARTS_DIR = ..

# Package up the chart files and move the tarball to the helm-charts repository
# Include a .helmignore file in your directory to define files to omit from the package.
package: init
	helm package -d $(HELM_CHARTS_DIR)/docs .

# This satifies the need for a test target but does nothing.
test:
	@:

init:
	@helm init -c > /dev/null

.PHONY: test package init
