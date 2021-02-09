#!/usr/bin/env python3
import collections
import itertools
import json
import operator
from typing import (
    Dict,
    Iterator,
    List,
)


def readlines(fpath: str) -> Iterator[str]:
    with open(fpath) as f:
        for line in f:
            yield line


def parse_titles(lines: Iterator[str]) -> List[str]:
    conversion_map = {
        'THE TRAGEDY OF ANTONY AND CLEOPATRA': 'ANTONY AND CLEOPATRA',
        'THE LIFE OF KING HENRY THE FIFTH': 'THE LIFE OF KING HENRY V',
        'THE TWO NOBLE KINSMEN': 'THE TWO NOBLE KINSMEN:',
        'TWELFTH NIGHT; OR, WHAT YOU WILL': 'TWELFTH NIGHT: OR, WHAT YOU WILL',
        'THE TRAGEDY OF OTHELLO, MOOR OF VENICE': 'OTHELLO, THE MOOR OF VENICE',
        'THE TRAGEDY OF MACBETH': 'MACBETH',
    }

    titles = []
    is_contents = False
    for line in lines:
        line = line.strip()
        if not line:
            continue
        if line == 'Contents':
            is_contents = True
            continue
        if line in titles:
            return titles
        if is_contents:
            title = conversion_map.get(line, line)
            titles.append(title)


def parse_works(titles: List[str], lines: Iterator[str]) -> Dict[str, str]:
    current_title = titles[0]
    titles_set = set(titles)
    work_map = collections.defaultdict(list)
    for line in lines:
        line_stripped = line.strip()

        if line_stripped == '* CONTENT NOTE (added in 2017) *':
            break

        if line_stripped in titles_set and line_stripped != current_title:
            current_title = line_stripped
            continue

        work_map[current_title].append(line)
    works = [{'title': title, 'content': '\n'.join(content)} for title, content in work_map.items()]
    works.sort(key=operator.itemgetter('title'))

    return works


def main():
    lines = readlines('./completeworks.txt')
    titles = parse_titles(lines)
    result = parse_works(titles, lines)

    with open('data.json', 'w') as f:
        json.dump(result, f, indent=4)


if __name__ == '__main__':
    main()
