# status-reporter
Provide Status Reports on a Server via HTTP

The administrators of a server need to be able to monitor its current state.
Tradtionally this is done by constantly producing a plain-text log file,
but nowadays we run severs in lightweight containers
which may not have very much disk space,
so running a constant log can be wasteful.

A better solution can be to keep the amount of logging to a minimum during normal running,
allow it to be increased when necessary
and provide a web service that produces a status report on demand.
All servers are different,
so the contents of the report sould be configurable.

Some logging can still be useful, particularly if the level can be turned up and down.
- POST /loglevel/n set log level to integer n
- GET /status get a status report

Each results in a function call on the server,
which should follow the StatusReporter interface.

The response to the status report call can be pre-formatted text, HTML or JSON.
The choice is made when the service is created.

Note that allowing arbitrary text in an HTML response introduces the risk of
an injection attack.
The server designer must control the contents of any status report,
ensuring that it does not quote verbatim
any data provided by a third party.
For example, the status report provided by the proxy in this repository
displays the contents of buffers that could potentially contain any text.
It guards against an injection attack by
pre-processing the buffers
and changing any '<' into '&amp;lt;'
and any '>' into '&amp;gt;'.