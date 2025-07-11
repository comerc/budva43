@who:user
@what:auto-answers
@why:automation

Feature: 11.AutoAnswers

  Background:
    Given исходный чат "<src_chat_id>" (<src_chat_name>)
    And исходное сообщение с вопросом

  Scenario Outline: Автоматические ответы на сообщения
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>)
    When пользователь отправляет исходное сообщение
    Then бот автоматически отвечает на сообщение

    Examples:
      | src_chat_id    | src_chat_name |
      | -1002641439846 | SRC PUB CH 1  |
      | -1002748936346 | SRC PUB CH 2  |
      | -1002792282007 | SRC PRV CH 1  |
      | -1002524362679 | SRC PRV CH 2  |
      | -1002736661856 | SRC PUB GRP 1 |
      | -1002781642357 | SRC PUB GRP 2 |
