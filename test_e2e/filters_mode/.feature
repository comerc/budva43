@who:user
@what:фильтр
@why:relevance

Feature: 04.FiltersMode

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Сообщение проходит фильтр include/exclude
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>)
    When пользователь отправляет сообщение с тегом "#ARK"
    Then сообщение появляется в целевом чате
    When пользователь отправляет сообщение с тегом "#УТРЕННИЙ_ОБЗОР"
    Then сообщение не появляется в целевом чате

    Examples:
      | dst_chat_id      | dst_chat_name   |
      | -1002667730628   | DST PUB CH 1    |
      | -1002877966922   | DST PUB CH 2    |
      | -1002473038431   | DST PRV CH 1    |
      | -1002641980237   | DST PRV CH 2    |
      | -1002866470933   | DST PUB GRP 1   |
      | -1002876400294   | DST PUB GRP 2   |
      | -4867965570      | DST GRP 1       |
      | -4913098869      | DST GRP 2       | 