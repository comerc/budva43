@who:user
@what:links
@why:cleanliness

Feature: 03.ReplaceMyselfLinks

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Ссылки на свои сообщения заменяются, внешние удаляются
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>) с replace-myself-links: run=true, delete-external=true
    When пользователь отправляет сообщение с ссылкой на своё сообщение
    Then ссылка заменяется на новую в целевом чате
    When пользователь отправляет сообщение с внешней ссылкой
    Then внешняя ссылка удаляется

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