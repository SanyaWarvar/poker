|EventType|EventMessage|Trigger|
|----|--------|----|
player_enter | player {{uuid}} enter the game | Вход в лобби нового игрока
game_started | game {{uuid}} started | Начало игры
players_stats | [ {id: uuid, balance: int} ] | В начале каждого раунда и после выплат в конце игры
new_round | new round started. Current round: {{int}} | В начале каждого раунда
get_cards | player {{uuid}} get cards: [ {{card}} ] | В начале пре-флоппа
community_cards | community cards: [ {{card}} ] | В начале флопа, терна, ривера
stop_game | game {{uuid}} has been stopped | В конце игры, когда завершился ривер и были произведены выплаты
win_all | player {{uuid}} win all pots with {{int}} total amount | Если все игроки, кроме одного, сбросили
win_pot | winners of pot {{int}} with {{int}} amount: [ {{uuid}} ] | В случае, если 2+ игрока не сбросили карты. Может быть ситуация, когда один банк делят несколько игроков, {{int}} указывает сколько досталось каждому
cant_ante | player {{uuid}} cant bet ante | Игроку не хватает баланса, чтобы поставить анте
get_ante | get ante: {{int}} | Сколько анте собрано
small_blind | player {{uuid}} bet {{int}} as small blind | В начале пре-флоппа
big_blind | player {{uuid}} bet {{int}} as big blind | В начале пре-флоппа
next_move | next move expect from {{uuid}} player | Когда любой игрок сделал ход - следующий в очереди получает оповещение
dealer | dealer is {{uuid}} | В начале пре-флоппа
bad_move | you cant check | Если игрок не может сделать чек
bad_move | raise must exceed the current bet by at least two times | Рейз обязан быть x2 от текущей ставки
bad_move | not enough money for this action | Не хватает денег, чтобы сделать raise. На call не влияет
bad_move | unexpected action | Если при отправке хода было отправлено что-то кроме check, call, fold, raise
can_do | player {{uuid}} can do call with {{int}} | Приходит сразу после next_move
can_do | player {{uuid}} can do check | Приходит сразу после next_move
do | player {{uuid}} do call with {{int}} amount | Приходит, когда какой то игрок сделал соответствующий ход
do | player {{uuid}} do raise with {{int}} amount | Приходит, когда какой то игрок сделал соответствующий ход
do | player {{uuid}} do check | Приходит, когда какой то игрок сделал соответствующий ход
do | player {{uuid}} do fold | Приходит, когда какой то игрок сделал соответствующий ход
***