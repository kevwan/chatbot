# chatbot

English | [简体中文](readme-cn.md)

`chatbot` is a conversational dialog engine build in Go which makes it possible to generate responses based on collections of known conversations. The language independent design of `chatbot` allows it to be trained to speak any language. It’s much faster than `ChatterBot`, our benchmark on 120 million conversations：

- ChatterBot takes 21s to answer
- chatbot takes 18ms to answer

## Project layout and command line tools

* bot

  Conversational dialog engine

* cli

  * train

    Train the given conversation data and generate corpus format file `.gob`

    * `-d` reads all `json` and `yaml` corpus files in the specified directory
    * `-i` read the specified `json` or `yaml` corpus files, splitting multiple files by commas
    * `-o` specify the output `.gob` file
    * `-m` print memory usage at regular intervals

  * ask

    An example question and answer command line tool

    * `-v` verbose
    * `-c` trained `.gob` file
    * `-t` data for several possible answers

## Data format

The data format can be provided via `yaml` or `json` files, refer to the format in `https://github.com/kevwan/chatterbot-corpus`. Roughly, it is as follows.

```yaml
categories:
- AI
- artificial intelligence
conversations:
- - What is AI?
  - Artificial Intelligence is the branch of engineering and science devoted to constructing machines that think.
- - Are you sentient?
  - Sort of.
```

## Example of a question and answer

```text
user: how are you?
bot: I am doing well, how about you?
user: I'm doing well, thanks!
bot: That is good to hear
```

## Acknowledgements

go-zero - https://github.com/zeromicro/go-zero

The `MapReduce` implementation of `go-zero in `core/mr` package gives `chatbot` a huge performance boost!

ChatterBot - https://github.com/gunthercox/ChatterBot

I was using `ChatterBot` at first, but it responds too slow, so I decided to implement it myself. Thanks to ChatterBot, great project!