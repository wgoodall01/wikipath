# wikipath [![Build Status](https://travis-ci.org/wgoodall01/wikipath.svg?branch=master)](https://travis-ci.org/wgoodall01/wikipath) [![Go Report Card](https://goreportcard.com/badge/github.com/wgoodall01/wikipath)](https://goreportcard.com/report/github.com/wgoodall01/wikipath) [![codecov](https://codecov.io/gh/wgoodall01/wikipath/branch/master/graph/badge.svg)](https://codecov.io/gh/wgoodall01/wikipath)
</h1>
Find the 'six degrees of wikipedia' between any two subjects, also known as the Wiki Game.

Pick any two Wikipedia articles. Starting at the first, click links until you reach the second. It's hard to do by hand, but also surprisingly challenging to automate.



## Examples

```
First Article  : Sodium Acetate
Second Article : Geneva Convention

      Sodium acetate  ->       53
  Geneva Conventions  <-      940

Searching for path... [done in 0.00s]
Path:  Sodium acetate > Textile > Bullet > Geneva Conventions
```

```

First Article  : List of role-playing video games: 2014 to 2015
Second Article : Barhait (Vidhan Sabha constituency)

List of role-playing video games: 2014 to 2015  ->      101
Barhait (Vidhan Sabha constituency)             <-        8

Searching for path... [done in 0.04s]
Path:  
    List of role-playing video games: 2014 to 2015 
  > Unrest (video game) 
  > India 
  > Bharatiya Janata Party 
  > Jharkhand Legislative Assembly 
  > Barhait (Vidhan Sabha constituency)
```

## How it works

Wikipath uses a bidirectional implementation of Dijkstra's algorithm to find the shortest path, via links clicked and redirects followed, between two Wikipedia articles.

The whole process follows these steps:

1. It starts with a full, compressed Mediawiki dump. You can get these from [dumps.wikimedia.org](https://dumps.wikimedia.org/).

1. The first pass, running `wikipath index`, converts that dump to a smaller, compressed binary format which only contains article titles and link destinations. This file is ~15x smaller, and much faster to parse, which causes a ~20x speedup in loading the articles into memory. It decompresses and parses every bzip stream in parallel for better performance, but this step still takes by far the longest.

1. When starting the program, with `wikipath start`, it will load that binary file into memory. Each article is allocated a struct of its title and an array of pointers to other articles.

1. Then, to build the index, it fills out the article structs. For each article, it will A) fill out an array of article pointers representing the articles it links to, and B) fill out another array of pointers to articles which link to it. The set of textual links are deleted to save memory.


1. To find the path between two articles, it will run Dijkstra's algorithm bidirectionally between the starting and ending articles. For each article, it will store the path from the start/end node at which the article was originally encountered. If the search encounters an article with a path coming from the opposite direction, it will terminate and return a result.
