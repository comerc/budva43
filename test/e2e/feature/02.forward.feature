@who:user
@what:forward
@why:forward

Feature: 02.Forward

  Background:
    Given будет пересылка - форвард

  Scenario Outline: Сообщение пересылается в целевые чаты
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    When пользователь отправляет сообщение в исходный чат
    Then пауза 10 сек.
    And сообщение в чате "-1002667730628" ("DST PUB CH 1")
    And сообщение в чате "-1002877966922" ("DST PUB CH 2")
    And сообщение в чате "-1002473038431" ("DST PRV CH 1")
    And сообщение в чате "-1002641980237" ("DST PRV CH 2")
    And сообщение в чате "-1002866470933" ("DST PUB GRP 1")
    And сообщение в чате "-1002876400294" ("DST PUB GRP 2")
    And сообщение в чате "-4867965570" ("DST GRP 1")
    And сообщение в чате "-4913098869" ("DST GRP 2")


    Examples:
      | src_chat_id    | src_chat_name |
      | -1002748936346 | SRC PUB CH 2  |
      | -1002524362679 | SRC PRV CH 2  |
      | -1002781642357 | SRC PUB GRP 2 |
