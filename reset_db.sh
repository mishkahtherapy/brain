#!/bin/bash
set -x
rm -f brain.db
sqlite3 brain.db < schema.sql
# sqlite3 brain.db < sample.sql