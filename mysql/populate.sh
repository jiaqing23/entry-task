#!/bin/bash
python3 datagen.py
mysql -uroot -p123456 < create.sql 
mysql -uroot -p123456 entry_task < insert_test.sql
mysql -uroot -p123456 entry_task < insert.sql