@who:user
@what:фильтр
@why:relevance

Feature: 04.1.filters_mode_exclude

  Scenario Outline: Сообщение не проходит фильтр (exclude) - пересылка копия
    Given будет пересылка - копия
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    # Это сообщение для сброса "EXCLUDE"
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4867965570)
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "EXCLUDE"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And нет сообщения в чате "DST PUB CHL 1" (1002667730628)
    And нет сообщения в чате "DST PRV CHL 1" (1002473038431)
    And нет сообщения в чате "DST PUB GRP 1" (1002866470933)
    And нет сообщения в чате "DST PRV GRP 1" (4867965570)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |

  Scenario Outline: Сообщение не проходит фильтр (exclude) - пересылка форвард
    Given будет пересылка - форвард
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    # Это сообщение для сброса "EXCLUDE"
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 2" (1002877966922)
    And сообщение в чате "DST PRV CHL 2" (1002641980237)
    And сообщение в чате "DST PUB GRP 2" (1002876400294)
    And сообщение в чате "DST PRV GRP 2" (4913098869)
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "EXCLUDE"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And нет сообщения в чате "DST PUB CHL 2" (1002877966922)
    And нет сообщения в чате "DST PRV CHL 2" (1002641980237)
    And нет сообщения в чате "DST PUB GRP 2" (1002876400294)
    And нет сообщения в чате "DST PRV GRP 2" (4913098869)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 2 | 1002748936346 |
      | SRC PRV CHL 2 | 1002524362679 |
      | SRC PUB GRP 2 | 1002781642357 |