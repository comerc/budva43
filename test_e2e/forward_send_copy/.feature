@who:user
@what:send-copy
@why:copy

Feature: 01.Forward.SendCopy

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Сообщение копируется в целевой чат
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>)
    When пользователь отправляет текстовое сообщение в исходный чат
    Then сообщение появляется в целевом чате как копия
    And сообщение содержит исходный текст
    But сообщение не содержит тегов приватности

    Examples:
      | dst_chat_id    | dst_chat_name |
      | -1002667730628 | DST PUB CH 1  |
      | -1002877966922 | DST PUB CH 2  |
      | -1002473038431 | DST PRV CH 1  |
      | -1002641980237 | DST PRV CH 2  |
      | -1002866470933 | DST PUB GRP 1 |
      | -1002876400294 | DST PUB GRP 2 |
      | -4867965570    | DST GRP 1     |
      | -4913098869    | DST GRP 2     |