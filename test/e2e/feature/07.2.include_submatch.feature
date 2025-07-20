Feature: 07.2.include_submatch

  Background:
    Given будет пересылка - форвард

  Scenario Outline: Сообщение проходит фильтр include-submatch
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "\$TSLA other"
    And будет текст "\$TSLA"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 2" (1002877966922)
    And сообщение в чате "DST PRV CHL 2" (1002641980237)
    And сообщение в чате "DST PUB GRP 2" (1002876400294)
    And сообщение в чате "DST PRV GRP 2" (4913098869)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 4 | 1002861761547 |
      | SRC PRV CHL 4 | 1002741670036 |
      | SRC PUB GRP 4 | 1002726501676 |
      | SRC PRV GRP 4 | 4834622030    |
