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

    cur.execute("INSERT INTO Reservations WITH RECURSIVE res(id,user,machine,start,end,ended,pxepath,nfsroot)AS(SELECT random(),(SELECT users.id from Users),(SELECT machines.id FROM Machines),abs(random()% (strftime('%s','2017-10-31 23:59'))),strftime('%s','2017-01-01 00:00') + abs(random()% (strftime('%s','2018-01-31 23:59') - strftime('%s','2017-01-01 00:00'))),NULL,abs(random()%10),abs(random()%10)UNION ALL SELECT random(),(SELECT users.id from Users),(SELECT machines.id FROM Machines),abs(random()% (strftime('%s','2017-01-31 23:59'))),strftime('%s','2017-01-01 00:00') + abs(random()% (strftime('%s','2018-01-31 23:59') - strftime('%s','2017-01-01 00:00'))),NULL,abs(random()%10),abs(random()%10)FROM res LIMIT  )select * from res;")

if con:

    con.close()

