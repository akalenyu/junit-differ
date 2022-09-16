# junit-differ

This is basically https://github.com/kubevirt/kubevirt/blob/46ffcbf1bacd3cdf1ae074d92dac74669b113c80/tools/junit-merger/junit-merger.go
that does ginkgo JUnit diffing instead of merging.

## Examples
Two XMLs included (real reports from kubevirt/kubevirt PRs), alongisde a result XML that was produced by the script.
- Build with `go build`
- Call with `./junit-differ-binary -o result.xml 8240.xml 8242.xml`
