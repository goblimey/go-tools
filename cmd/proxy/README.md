# proxy

A Man In The Middle (MITM) Proxy with Status Reporting, written in Go

This is a rework of an original version by Staaldraad.
The original is available [here](https://github.com/staaldraad/tcpprox)
and described [here](https://staaldraad.github.io/2016/12/11/tcpprox/).

This version uses the status-reporter to control logging levels and provide status reports.

## Installation

     git clone https://github.com/goblimey/go-tools.git

     go install github.com/goblimey/go-tools/proxy


This produces a program proxy.


## Running the proxy

Running on server {servername},
receiving HTTP requests on port 2102
and passing them onto a server on localhost port 2101.
Taking control requests on port 4001
with logging to initially quiet (log level 0):

    proxy -p 2102 -r localhost:2101 -l {servername} -ca {servername} -cp 4001 -q &

The log file will be ./proxy.{yyyy-mm-dd}.log, which will roll over every day.


## Log Level

Set the log level to 1:

    curl -X POST {servername}:{port}/status/loglevel/1

where {servername} and {port} are the servername and port.

The log level value is 0-255.  0 turns logging off.
Any value above 0 turns logging on.


## Status Report

To produce a status report:

    curl {servername}:{port}/status/report

where {servername} and {port} are the servername and port.
 
```
Status
Last Client Buffer

From Client [3]: Thu Dec 12 17:42:33 2019

00000000  f8 00 05 54 e5 75 f5 df  55 b3 e9 94 b5 4f 58 00  |...T.u..U....OX.|
00000010  00 00 00 00 00 00 00 00  00 00 00 e1 50 cc d3 00  |............P...|
00000020  8a 43 c0 00 91 c6 5a a2  00 00 41 c0 f0 00 00 00  |.C....Z...A.....|
00000030  00 00 20 80 00 00 7f ff  a6 22 a2 20 25 27 24 25  |.. ......". %'$%|
00000040  9c ad 40 1e c2 26 9b 76  ec bc 87 56 0f 54 4f bc  |..@..&.v...V.TO.|
00000050  9c 70 05 9f f7 7e 37 bc  65 87 2f 0e 03 bc 43 6c  |.p...~7.e./...Cl|
00000060  79 c8 e3 98 27 2b 4e 2d  01 d8 84 07 72 50 4b 93  |y...'+N-....rPK.|
00000070  e1 11 b6 80 46 04 08 58  77 e7 6e bf a5 0b 82 1e  |....F..Xw.n.....|
00000080  40 09 af 1f b8 b2 3e b6  61 87 7c 1e 1e b6 d8 74  |@.....>.a.|....t|
00000090  39 01 eb 02 ff ff ff ff  ff d7 ff ff 80 00 47 35  |9.............G5|
000000a0  f7 d9 86 39 4d 3b af 59  74 33 80 e2 c4 5c d3 00  |...9M;.Yt3...\..|
000000b0  8a 44 60 00 61 9a 05 e2  00 00 14 40 00 e1 40 00  |.D`.a......@..@.|
000000c0  00 00 20 01 00 00 7f ef  ab 28 28 a4 2c 2c 29 2d  |.. ......((.,,)-|
000000d0  c5 d7 19 8e 38 4a ed ef  27 03 71 cd e3 5e 77 d0  |....8J..'.q..^w.|
000000e0  ee 61 75 b2 ef 21 88 c4  34 95 4b 2a 9c 4e 54 2d  |.au..!..4.K*.NT-|
000000f0  70 5b a0 e2 21 b4 be 2f  32 f9 64 40 3c f7 90 fc  |p[..!../2.d@<...|
00000100  15 43 1a 33 0c 80 50 10  6c 50 62 64 02 dc d1 0c  |.C.3..P.lPbd....|
00000110  7c 1c 26 ca 00 39 db 81  1d f6 03 c3 90 0e 6d 4f  ||.&..9........mO|
00000120  ff ff ff ff ff ff ff 00  01 5e 57 eb 56 38 6d 4e  |.........^W.V8mN|
00000130  0f db a4 f8 00 00 00 00  00 00 00 dc 46 4b d3 00  |............FK..|
00000140  dc 44 90 00 61 9a 05 e2  00 00 14 40 00 e1 40 00  |.D..a......@..@.|
00000150  00 00 20 01 00 00 7f ef  ab 28 28 a4 2c 2c 29 2d  |.. ......((.,,)-|
00000160  80 00 00 00 45 d7 19 8e  38 4a ed ef 27 03 02 c9  |....E...8J..'...|
00000170  fa 97 b7 e0 33 83 b7 fe  8f d2 20 bf f1 cc ff 1a  |....3..... .....|
00000180  f4 1d f4 31 dc b8 97 5b  59 77 8a 06 22 90 86 8a  |...1...[Yw.."...|
00000190  15 4b 59 54 db 13 94 c8  5a e0 85 ba 28 71 09 86  |.KYT....Z...(q..|
000001a0  d1 af c5 e6 3f cb 22 08  79 ef 10 7e 0a a8 63 46  |....?.".y..~..cF|
000001b0  60 64 02 78 20 d8 98 31  31 f0 5b 9a 18 63 e0 d0  |`d.x ..11.[..c..|
000001c0  4d 93 f0 1c ed b8 23 be  c8 1e 1c 90 1c da 7c 93  |M.....#.......|.|
000001d0  25 4e 13 9d 15 46 4d 93  74 73 1d bd d4 1d 07 0c  |%N...FM.ts......|
000001e0  43 40 00 2b 0c 82 f0 d4  2a 0c 43 00 d8 29 0c 01  |C@.+....*.C..)..|
000001f0  f0 b4 34 09 c3 00 31 2a  62 17 b5 2f 6a 5e df 5d  |..4...1*b../j^.]|
00000200  ca d2 f1 25 e4 28 50 50  37 41 df 00 5e 06 00 16  |...%.(PP7A..^...|
00000210  00 14 80 00 00 00 00 00  00 00 00 00 00 18 04 59  |...............Y|
00000220  d3 00 3f 46 40 00 61 99  2b 22 00 00 00 00 16 12  |..?F@.a.+"......|
00000230  00 00 00 00 20 00 00 00  7d 5d 41 3d 41 4a 58 58  |.... ...}]A=AJXX|
00000240  e8 1d eb 29 cc 3c 44 b7  26 88 5e 40 43 fe 62 e0  |...).<D.&.^@C.b.|
00000250  84 5f 7f e5 68 a0 10 fb  40 3b 7f 57 ff f8 24 b7  |._..h...@;.W..$.|
00000260  1b ab 05 a9 6f d3 00 5f  46 70 00 61 99 2b 20 00  |....o.._Fp.a.+ .|
00000270  00 00 00 16 12 00 00 00  00 20 00 00 00 7d 5d 41  |......... ...}]A|
00000280  3d 41 48 00 00 25 85 8e  81 de b2 9f 59 80 89 f9  |=AH..%......Y...|
00000290  60 07 80 8b f3 0f 14 89  69 32 68 44 2f 21 41 0f  |`.......i2hD/!A.|
000002a0  63 cc 5c 18 22 fb f7 ca  d1 3c 08 7d a0 07 6f dd  |c.\."....<.}..o.|
000002b0  45 93 a7 49 26 14 04 81  68 62 17 05 61 b4 c0 2b  |E..I&...hb..a..+|
000002c0  be 01 c2 cc c6 64 40 bd  6b a9 d3 00 08 4c e0 00  |.....d@.k....L..|
000002d0  8a 00 00 00 00 a8 f7 2a                           |.......*|



Last Server Buffer

To Server [3]: Sat Dec 7 19:31:32 2019

00000000  4f 4b 0d 0a                                       |OK..|
```
