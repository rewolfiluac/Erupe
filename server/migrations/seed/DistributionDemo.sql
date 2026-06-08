BEGIN;

-- HR5 guide reward from official server screenshots.
-- Equipment item types: 0=legs, 1=head, 2=chest, 3=arms, 4=waist.
INSERT INTO distribution (type, event_name, description, times_acceptable, min_hr, max_hr)
VALUES (1, 'HR5突破の褒賞品', $desc$~C05このイベント報酬を受け取るには、
下記の条件を満たしている必要があります。

HR5を突破したハンターにギルドからの褒賞品です！

剛種武器は、武具工房の親方に話しかけて
「武器加工」＞「剛種武器生産」から作りたい武器種を
選択してください。

【達成条件】
・HR5以上

【プレゼントアイテム】
・コルーデFYシリーズ一式
・剛種武器「輝界白竜シリーズ」の生産素材
・剛力珠、ポルタチケット桜、ＧＰ交換券

詳細は、公式サイトでご確認ください。$desc$, 1, 5, 999);
INSERT INTO distribution_items (distribution_id, item_type, item_id, quantity) VALUES
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 1, 7382, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 1, 7389, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 2, 6691, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 2, 6698, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 3, 6684, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 3, 6691, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 4, 6838, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 4, 6845, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 0, 6684, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 0, 6691, 1),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 910, 8),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1472, 90),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 8174, 15),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 8175, 25),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 8143, 36),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 8146, 6),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 8142, 36),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1426, 20),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1423, 4),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1411, 5),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1420, 8),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 2209, 6),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1410, 8),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1408, 6),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1405, 8),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 1417, 4),
((SELECT id FROM distribution ORDER BY id DESC LIMIT 1), 7, 13190, 250);

END;
