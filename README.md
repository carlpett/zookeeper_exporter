## zookeeper_exporter [![CircleCI](https://circleci.com/gh/carlpett/zookeeper_exporter.svg?style=shield)](https://circleci.com/gh/carlpett/zookeeper_exporter)

A very simple prometheus exporter for zookeeper 3.4+. 

### Limitations
Due to the type of data exposed by Zookeeper's `mntr` command, it currently resets Zookeeper's internal statistics every time it is scraped. This makes it unsuitable for having multiple parallel scrapers.
