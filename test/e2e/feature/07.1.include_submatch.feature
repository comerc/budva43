Feature: 07.1.include_submatch

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение проходит фильтр include-submatch
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "\$TSLA other"
    And будет текст "\$TSLA"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4897079215)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 3 | 1002828900048 |
      | SRC PRV CHL 3 | 1002557642474 |
      | SRC PUB GRP 3 | 1002531681006 |
      | SRC PRV GRP 3 | 4836830199    |

