Feature: 14.translate

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение копируется в целевой чат с переводом
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "awesome message"
    And будет текст "отличное сообщение"
    When пользователь отправляет сообщение
    Then ожидание 15 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4897079215)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 5 | 1002721762679 |
      | SRC PRV CHL 5 | 1002707662151 |
      | SRC PUB GRP 5 | 1002646245904 |
      | SRC PRV GRP 5 | 4898677378    |
