Feature: 12.copy_once

  # Background:
  #   Given исходный чат "<src_chat_name>" (<src_chat_id>)
  #   And режим "copy-once: true"

  # Scenario Outline: Сообщение копируется только один раз
  #   Given целевой чат "<dst_chat_id>" (<dst_chat_name>)
  #   When пользователь отправляет исходное сообщение
  #   Then сообщение появляется в целевом чате
  #   And сообщение копируется только один раз

  Background:
    Given будет пересылка - копия

  Scenario Outline: Сообщение копируется только один раз
    Given исходный чат "<src_chat_name>" (<src_chat_id>)
    And сообщение с текстом "some text"
    And будет текст "some text"
    When пользователь отправляет сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    # And сообщение в чате "DST PRV CHL 1" (1002473038431)
    # And сообщение в чате "DST PUB GRP 1" (1002866470933)
    # And сообщение в чате "DST PRV GRP 1" (4867965570)
    Given сообщение с текстом "some other text"
    And будет текст "some text"
    When пользователь редактирует сообщение
    Then ожидание 10 сек.
    And сообщение в чате "DST PUB CHL 1" (1002667730628)
    # And сообщение в чате "DST PRV CHL 1" (1002473038431)
    # And сообщение в чате "DST PUB GRP 1" (1002866470933)
    # And сообщение в чате "DST PRV GRP 1" (4867965570)

    Examples:
      | src_chat_name | src_chat_id   |
      | SRC PUB CHL 1 | 1002641439846 |
# | SRC PRV CHL 1 | 1002792282007 |
# | SRC PUB GRP 1 | 1002736661856 |
# | SRC PRV GRP 1 | 4832061506    |

