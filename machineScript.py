#!/usr/bin/python
# -*- coding: utf-8 -*-

import sqlite3 as lite
import sys
import datetime
import random
from datetime import date



con = lite.connect('clowder.db')

with con:

    cur = con.cursor()

    cur.execute("INSERT INTO Machines WITH RECURSIVE mch(id, name, arch, microarch, cores, memory)AS(SELECT random(),abs(random()% 100),abs(random()%10),abs(random()%10),abs(random()%12),abs(random() %12)UNION ALL SELECT random(),abs(random() % 100),abs(random()%10),abs(random()%10),abs(random()%12),abs(random()%12)FROM mch LIMIT 5 ) select * from mch;")
   
if con:

    con.close()


