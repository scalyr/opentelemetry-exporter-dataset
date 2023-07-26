#!/usr/bin/env bash
# Script which checks that GNU compatible sed, date and other binaries are available (e.g. if
# running on OS X).
# A lot of our bash scripts rely on GNU tool specific functionality which is not available in BSD
# tools which are available out of the box on OS X.

# BSD date doesn't support N modifier for nanoseconds and simply returns N string value instead of
# actual nanoseconds.
DATE_OUTPUT=$(date "+%s%N")

if [[ "${DATE_OUTPUT}" == *"N" ]]; then
    echo "❌  date command returned invalid value which means it's likely using BSD and not GNU versions of the tool."
    echo "❌  If you are running on OS X make sure to installs coreutils package via homebrew or similar and make sure PATH is configured correctly."
    echo "❌  Keep in mind that on m1/m2 default homebrew bin path is different than on intel mackbooks."
    echo "❌  See https://github.com/scalyr/opentelemetry-exporter-dataset#dependencies for more details."
    exit 1
fi
