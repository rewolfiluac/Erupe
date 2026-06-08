-- Align the HR5 guide reward with official server screenshots.
-- The first ten entries are Coord FY equipment, awarded directly.
-- The remaining entries are the item rewards shown on pages 2 and 3.
DO $$
DECLARE
    dist_id INTEGER;
BEGIN
    SELECT id INTO dist_id
    FROM distribution
    WHERE type = 1
      AND character_id IS NULL
      AND event_name IN ('HR5 Welcome Pack', 'HR5突破の褒賞品')
    ORDER BY id
    LIMIT 1;

    IF dist_id IS NULL THEN
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

詳細は、公式サイトでご確認ください。$desc$, 1, 5, 999)
        RETURNING id INTO dist_id;
    ELSE
        UPDATE distribution
        SET event_name = 'HR5突破の褒賞品',
            description = $desc$~C05このイベント報酬を受け取るには、
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

詳細は、公式サイトでご確認ください。$desc$,
            times_acceptable = 1,
            min_hr = 5,
            max_hr = 999
        WHERE id = dist_id;

        DELETE FROM distribution_items WHERE distribution_id = dist_id;
        DELETE FROM distributions_accepted WHERE distribution_id = dist_id;
    END IF;

    INSERT INTO distribution_items (distribution_id, item_type, item_id, quantity) VALUES
    (dist_id, 1, 7382, 1),
    (dist_id, 1, 7389, 1),
    (dist_id, 2, 6691, 1),
    (dist_id, 2, 6698, 1),
    (dist_id, 3, 6684, 1),
    (dist_id, 3, 6691, 1),
    (dist_id, 4, 6838, 1),
    (dist_id, 4, 6845, 1),
    (dist_id, 0, 6684, 1),
    (dist_id, 0, 6691, 1),
    (dist_id, 7, 910, 8),
    (dist_id, 7, 1472, 90),
    (dist_id, 7, 8174, 15),
    (dist_id, 7, 8175, 25),
    (dist_id, 7, 8143, 36),
    (dist_id, 7, 8146, 6),
    (dist_id, 7, 8142, 36),
    (dist_id, 7, 1426, 20),
    (dist_id, 7, 1423, 4),
    (dist_id, 7, 1411, 5),
    (dist_id, 7, 1420, 8),
    (dist_id, 7, 2209, 6),
    (dist_id, 7, 1410, 8),
    (dist_id, 7, 1408, 6),
    (dist_id, 7, 1405, 8),
    (dist_id, 7, 1417, 4),
    (dist_id, 7, 13190, 250);
END $$;
