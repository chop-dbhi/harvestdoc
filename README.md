# Harvest Doc

Command line tool to generate a CSV file of Harvest fields and concepts.

## Install

```
go install github.com/chop.edu/harvestdoc
```

## Example

This downloads the fields from the public Harvest demo.

```
harvestdoc http://harvest.research.chop.edu/demo/api/ > demo.csv
```
