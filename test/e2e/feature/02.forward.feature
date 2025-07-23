Feature: 02.forward

  Background:
    Given будет пересылка - форвард

  Scenario Outline: Сообщение пересылается в целевой чат
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 2" (1002877966922)
    And сообщение в чате "DST PRV CHL 2" (1002641980237)
    And сообщение в чате "DST PUB GRP 2" (1002876400294)
    And сообщение в чате "DST PRV GRP 2" (4913098869)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 2 | 1002748936346 |
      | SRC PRV CHL 2 | 1002524362679 |
      | SRC PUB GRP 2 | 1002781642357 |
      | SRC PRV GRP 2 | 4977845927    |
