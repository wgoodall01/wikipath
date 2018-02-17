#!/usr/bin/env bash

aria2c -c -x16 -s16 -i - <<EOF
https://dumps.wikimedia.org/simplewiki/20180101/simplewiki-20180101-pages-articles-multistream.xml.bz2
	out=simple.xml.bz2
	checksum=sha-1=e1dbcfd1fdf421572f68638ab5230e675295e215

https://dumps.wikimedia.org/simplewiki/20180101/simplewiki-20180101-pages-articles-multistream-index.txt.bz2
	out=simple-index.txt.bz2
	checksum=sha-1=df180b0934d55b32859ac3216833dcc8f0c82bf9

https://dumps.wikimedia.org/enwiki/20180101/enwiki-20180101-pages-articles-multistream.xml.bz2
	out=complete.xml.bz2
	checksum=sha-1=0a4bee239be61a8ef77c2e72d9b4546863c57264

https://dumps.wikimedia.org/enwiki/20180101/enwiki-20180101-pages-articles-multistream-index.txt.bz2
	out=complete-index.txt.bz2
	checksum=sha-1=8de15c617180d9cb8798d8bb60b2090894ef6c42
EOF
