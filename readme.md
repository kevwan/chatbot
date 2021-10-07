# chatbot

[English](readme-en.md) | 简体中文

## 项目说明

`chatbot` 是一个通过已知对话数据集快速生成回答的 Go 问答引擎。比 [ChatterBot](https://github.com/gunthercox/ChatterBot) 快非常多，我们在1.2亿对话上的对比结果是：[ChatterBot](https://github.com/gunthercox/ChatterBot) 回答需要21秒，chatbot 只需要18毫秒。

* bot

  问答引擎

* cli
  
  * train
  
    训练给定的问答数据并生成 `.gob` 文件
  
    * `-d` 读取指定目录下所有 `json` 和 `yaml` 语料文件
    * `-i` 读取指定的 `json` 或 `yaml` 语料文件，多个文件用逗号分割
    * `-o` 指定输出的 `.gob` 文件
    * `-m` 定时打印内存使用情况
  
  * ask
  
    一个示例的问答命令行工具
  
    * `-v` verbose
    * `-c` 训练好的 `.gob` 文件
    * `-t` 数据几个可能的答案

## 数据格式

数据格式可以通过 `yaml` 或者 `json` 文件提供，参考 `https://github.com/kevwan/chatterbot-corpus` 里的格式。大致如下：

```yaml
categories:
- AI
conversations:
- - 什么是ai
  - 人工智能是工程和科学的分支,致力于构建具有思维的机器。
- - 你是什么语言编写的
  - Python
- - 你听起来像机器
  - 是的,我受到造物者的启发
- - 你是一个人工智能
  - 那是我的名字。
```

## 问答示例

```text
user: 在吗？
bot: 在
user: 在干嘛呢？
bot: 看电视
user: 看啥电视呀
bot: 活色生香
user: 很好看吗？
bot: 特搞笑
user: 你在哪里呀？
bot: 家里
user: 家里就你一个人嘛？
bot: 我喜欢一个人玩
user: 那我过来找你？
bot: 不可以，乖乖上班去
```

## 致谢

ChatterBot - https://github.com/gunthercox/ChatterBot

最早我是使用 [ChatterBot](https://github.com/gunthercox/ChatterBot).的，但由于回答太慢，所有后来只能自己实现了，感谢 [ChatterBot](https://github.com/gunthercox/ChatterBot)，非常棒的项目！
