#!/usr/bin/env bash

aria2c -x16 -s16 -i - <<EOF
https://dumps.wikimedia.org/simplewiki/20180101/simplewiki-20180101-pages-articles-multistream.xml.bz2
	out=simple.xml.bz2
	checksum=sha-1=e1dbcfd1fdf421572f68638ab5230e675295e215

https://dumps.wikimedia.org/simplewiki/20180101/simplewiki-20180101-pages-articles-multistream-index.txt.bz2
	out=simple-index.txt.bz2
	checksum=sha-1=df180b0934d55b32859ac3216833dcc8f0c82bf9
EOF

echo "Unpacking..."
bzip2 -dk simple.xml.bz2
bzip2 -dk simple-index.txt.bz2

