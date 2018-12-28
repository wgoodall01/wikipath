#!/usr/bin/env bash

ID="20181220" # https://dumps.wikimedia.org/enwiki/{ID}/

# cd to script location
cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

aria2c -c -x16 -s16 -i - <<EOF
https://dumps.wikimedia.org/simplewiki/${ID}/simplewiki-${ID}-pages-articles-multistream.xml.bz2
	out=simple.xml.bz2

https://dumps.wikimedia.org/simplewiki/${ID}/simplewiki-${ID}-pages-articles-multistream-index.txt.bz2
	out=simple-index.txt.bz2

https://dumps.wikimedia.org/enwiki/${ID}/enwiki-${ID}-pages-articles-multistream.xml.bz2
	out=complete.xml.bz2

https://dumps.wikimedia.org/enwiki/${ID}/enwiki-${ID}-pages-articles-multistream-index.txt.bz2
	out=complete-index.txt.bz2
EOF

# maybe checksums in future:
#	checksum=sha-1=8de15c617180d9cb8798d8bb60b2090894ef6c42
