Feature: 13.2.indelible_f

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение удаляется из целевого чата
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    # "INCLUDE" побочно, чтобы выполнить forward-rules
    And сообщение с текстом "INCLUDE"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4897079215)
    Given сброс проверок
    When пользователь удаляет сообщение
    Then ожидание 10 сек.
    And нет сообщения в чате "DST PUB CHL 1" (1002667730628)
    And нет сообщения в чате "DST PRV CHL 1" (1002473038431)
    And нет сообщения в чате "DST PUB GRP 1" (1002866470933)
    And нет сообщения в чате "DST PRV GRP 1" (4897079215)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 3 | 1002828900048 |
      | SRC PRV CHL 3 | 1002557642474 |
      | SRC PUB GRP 3 | 1002531681006 |
      | SRC PRV GRP 3 | 4836830199    |

