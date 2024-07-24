#!/usr/bin/env bash

# build example from README.md
d=$(dirname "$0");
cd "${d}/../examples/readme" || exit 1

pwd;

# build example
rm -rfv readme
go build -o readme
ls -l readme

# try different domains
# https://badssl.com
for config in https://app.scalyr.com,buffers https://expired.badssl.com,x509 https://wrong.host.badssl.com,x509 https://self-signed.badssl.com,x509 https://sentinelone.com,parse; do
  url=$(echo ${config} | cut -f1 -d,);
  expectedPanic=$(echo ${config} | cut -f2 -d,);
  domain=$(echo ${url} | cut -f3 -d/)
  logFile=/tmp/test-ssh-${domain}.log;

  echo "Trying URL: ${url} (${domain}) and expects panic ${expectedPanic}";
  echo "Logs: ${logFile}";
  export SCALYR_SERVER="${url}";
  export SCALYR_WRITELOG_TOKEN="aaa";
  ./readme 2>&1 | tee ${logFile};
  status=$?

  echo "Url: ${url} (${domain}); Status: ${status}; Expected: ${expectedPanic}";
  grep panic ${logFile};
  if [ $( grep panic ${logFile} | grep -c "${expectedPanic}" ) == 0 ]; then
    echo "Expected panic ${expectedPanic} was not found :(";
    exit 1;
  else
    echo "Expected panic ${expectedPanic} was found :)";
  fi;
done;

echo "Everything is working as expected"
