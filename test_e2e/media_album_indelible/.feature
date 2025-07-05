@who:user
@what:album_indelible
@why:indelibility

Feature: 13.MediaAlbumIndelible

  Background:
    Given исходный чат "-1002641439846" (SRC PUB CH 1)

  Scenario Outline: Медиаальбом нельзя удалить из целевого чата
    Given целевой чат "<dst_chat_id>" (<dst_chat_name>) с indelible
    When пользователь отправляет медиаальбом в исходный чат
    Then альбом появляется в целевом чате и не может быть удалён

    Examples:
      | dst_chat_id    | dst_chat_name |
      | -1002667730628 | DST PUB CH 1  |
      | -1002877966922 | DST PUB CH 2  |