@who:user
@what:forward
@why:forward

Feature: 02.forward

  Background:
    Given будет пересылка - форвард

  Scenario Outline: Сообщение пересылается в целевые чаты
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    When пользователь отправляет сообщение
    Then пауза 10 сек.
    And сообщение в чате "-1002877966922" (DST PUB CHL 2)
    And сообщение в чате "-1002641980237" (DST PRV CHL 2)
    And сообщение в чате "-1002876400294" (DST PUB GRP 2)
    And сообщение в чате "-4913098869" (DST PRV GRP 2)


    Examples:
      | src_chat_id    | src_chat_name |
      | -1002748936346 | SRC PUB CHL 2 |
      | -1002524362679 | SRC PRV CHL 2 |
      | -1002781642357 | SRC PUB GRP 2 |
