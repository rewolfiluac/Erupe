-- Remove the client-provided event reward condition notice from the stored HR5 description.
UPDATE distribution
SET description = $desc$~C05HR5を突破したハンターにギルドからの褒賞品です！

剛種武器は、武具工房の親方に話しかけて
「武器加工」＞「剛種武器生産」から作りたい武器種を
選択してください。

【達成条件】
・HR5以上

【プレゼントアイテム】
・コルーデFYシリーズ一式
・剛種武器「輝界白竜シリーズ」の生産素材
・剛力珠、ポルタチケット桜、ＧＰ交換券

詳細は、公式サイトでご確認ください。$desc$
WHERE type = 1
  AND event_name = 'HR5突破の褒賞品'
  AND character_id IS NULL;
