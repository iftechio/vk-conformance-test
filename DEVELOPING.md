# 开发指南

在使用 virtual-kubelet 的过程中遇到不符合预期的行为，可以写一个最小的复现方法加到 [testcases](./testcases) 中 (可参考 [env100.go](./testcases/env100.go)) ，方便跟云厂商的开发反馈。

[test suite](./suite/suite.go) 中提供了些在测试里使用的辅助函数
