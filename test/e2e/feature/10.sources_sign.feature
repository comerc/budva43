@who:user
@what:sources-sign
@why:identification

Feature: 10.sources_sign

  Background:
    Given будет пересылка - копия
    And будет подпись

  Scenario Outline: Вставка подписи источника
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    When пользователь отправляет сообщение
    Then пауза 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    And сообщение в чате "DST PRV CHL 1" (1002473038431)
    And сообщение в чате "DST PUB GRP 1" (1002866470933)
    And сообщение в чате "DST PRV GRP 1" (4867965570)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
      | SRC PRV CHL 1 | 1002792282007 |
      | SRC PUB GRP 1 | 1002736661856 |
