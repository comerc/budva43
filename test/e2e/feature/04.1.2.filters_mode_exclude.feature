Feature: 04.1.2.filters_mode_exclude

  Background:
    Given будет пересылка - форвард

  Scenario Outline: Сообщение не проходит фильтр (exclude)
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    # Это сообщение для сброса "EXCLUDE"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 2" (1002877966922)
    And сообщение в чате "DST PRV CHL 2" (1002641980237)
    And сообщение в чате "DST PUB GRP 2" (1002876400294)
    And сообщение в чате "DST PRV GRP 2" (4913098869)
    # Переназначаем исходный чат - новый nanoid
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "EXCLUDE other"
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
      | SRC PRV GRP 2 | 4977845927    |
