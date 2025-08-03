Feature: 04.1.1.filters_mode_exclude

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение не проходит фильтр (exclude)
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    # Это сообщение для сброса "EXCLUDE"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4897079215)
    # Переназначаем исходный чат - новый nanoid
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "EXCLUDE other"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And нет сообщения в чате "DST PUB CHL 1" (1002667730628)
    And нет сообщения в чате "DST PRV CHL 1" (1002473038431)
    And нет сообщения в чате "DST PUB GRP 1" (1002866470933)
    And нет сообщения в чате "DST PRV GRP 1" (4897079215)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
      | SRC PRV GRP 1 | 4832061506    |
