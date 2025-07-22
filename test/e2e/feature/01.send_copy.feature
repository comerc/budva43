Feature: 01.send_copy

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение копируется в целевой чат
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4867965570)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
      | SRC PRV GRP 1 | 4832061506    |
