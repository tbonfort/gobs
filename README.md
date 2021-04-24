# Gobs is a tiny worker pool
[![Go Reference](https://pkg.go.dev/badge/github.com/tbonfort/gobs.svg)](https://pkg.go.dev/github.com/tbonfort/gobs)
[![License](https://img.shields.io/github/license/tbonfort/gobs.svg)](https://github.com/tbonfort/gobs/blob/main/LICENSE)
[![Build Status](https://github.com/tbonfort/gobs/workflows/build/badge.svg?branch=main&event=push)](https://github.com/tbonfort/gobs/actions?query=workflow%3Abuild+event%3Apush+branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/tbonfort/gobs/badge.svg?branch=main)](https://coveralls.io/github/tbonfort/gobs?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/tbonfort/gobs)](https://goreportcard.com/report/github.com/tbonfort/gobs)

Package gobs implements a simple job queue where each individual job is run
concurrently in its own goroutine while ensuring that no more than a given number
of jobs can be ran at a time. It provides methods to ensure all jobs have been
completed, and to capture errors.
