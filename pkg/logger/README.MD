# Logger

A facility that logs to the standard output channel
at various levels of detail.Level 0 disables logging.
It's assumed that level 1 sets minimal logging and
a bigger level number increases the detail.

    log = logger.MakeLogger()

creates a logger.

    log.SetLogLevel(0)
  
sets the level.

    fmt.Fprintf(log, "listening on %s\n", fmt.Sprint(config.Localhost, ":", config.Localport))

writes to the log channel.  This only has any effect if the level is greater than zero.