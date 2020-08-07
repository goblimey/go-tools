## SwitchWriter

The switch writer is a managed io.Writer that provides a destination stream
that can be changed dynamically.  The destination stream can be disabled,
in which case all writes to it are silently discarded.

The switch writer demo shows how it can be used.