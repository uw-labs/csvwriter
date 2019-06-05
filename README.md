csvwriter
=========

csvwriter is a simple CSV writer based on code in `encoding/csv` in the standard library.

The main differences from `encoding/csv` are:

* Takes []byte as inputs instead of string
* Prioritises performance over flexibility

