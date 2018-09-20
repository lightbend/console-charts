#!/usr/bin/env bash

# check the e2e result file
# the pass count should not be less then $EXPECTED_PASS
# the fail count should not be greater then $EXPECTED_FAIL

E2E_RESULT_FILE=results.txt

EXPECTED_PASS=8
EXPECTED_FAIL=4

# RESULT is last non-empty line | remove beginning to time
RESULT=`awk '/./{line=$0} END{print line}' $E2E_RESULT_FILE | sed -n 's/^.*:[0-9][0-9]//p'`
TESTS=`echo  $RESULT | awk '{ print $1 }'`
PASS=`echo  $RESULT | awk '{ print $2 }'`
FAIL=`echo  $RESULT | awk '{ print $3 }'`
PENDING=`echo  $RESULT | awk '{ print $4 }'`
SKIP=`echo  $RESULT | awk '{ print $5 }'`

echo "==================="
echo "parsed test results"
echo "TESTS=$TESTS"
echo "PASS=$PASS"
echo "FAIL=$FAIL"
echo "PENDING=$PENDING"
echo "SKIP=$SKIP"
echo ""


if [[ "$PASS" -lt "$EXPECTED_PASS" ]]
then
  echo "pass count $PASS should not be less then $EXPECTED_PASS"
  exit -1
fi

if [[ "$FAIL" -gt "$EXPECTED_FAIL" ]]
then
  echo "fail count $FAIL should not be greater then $EXPECTED_FAIL"
  exit -1
fi
