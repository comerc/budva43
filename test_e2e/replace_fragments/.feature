@who:user
@what:replace-fragments
@why:replacement

Feature: 08.ReplaceFragments

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Фрагменты текста заменяются в целевом чате
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>) с replace-fragments
    When пользователь отправляет сообщение с текстом "<from>"
    Then сообщение появляется в целевом чате с текстом "<to>"

    Examples:
      | dst_chat_id      | dst_chat_name   | from   | to     |
      | -1002667730628   | DST PUB CH 1    | hello  | 12345  |
      | -1002877966922   | DST PUB CH 2    | world  | 67890  | 