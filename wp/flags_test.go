package wikipath

import "flag"

var wikiArchivePath = flag.String("archivePath", "./wikis/simple.xml", "Path to the wiki dump, as an .xml file.")
var wikiArchiveBzipPath = flag.String("bzipPath", "./wikis/simple.xml.bz2", "Path to the wiki dump, as an .xml.bz2 file.")
var wikiArchiveIndexPath = flag.String("indexPath", "./wikis/simple-index.txt", "Path to the index, as a .txt")
var WikiArchiveIndexBzipPath = flag.String("bzipIndexPath", "./wikis/simple-index.txt.bz2", "Path to the index, as a .txt.bz2")
var wpindexPath = flag.String("wpindex", "./wikis/simple.wpindex", "Path to *.wpindex file")
