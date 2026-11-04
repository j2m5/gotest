-- Authored enemy concepts; not a port of legacy battle formulas or quest NPCs.
INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('underworld_scavenger','Мародёр нижних уровней','Промышляет кражами среди жилых переходов и заброшенных складов. Привык нападать там, где трудно позвать на помощь.','humanoid'),
 ('gang_enforcer','Бандит-вышибала','Силой удерживает проходы и торговые точки под контролем своей банды. Предпочитает тесные переулки открытым площадкам.','humanoid'),
 ('hired_sharpshooter','Наёмный стрелок','Устраивает засады на мостках и технических балконах. Хорошо знает укрытия городских нижних ярусов.','humanoid'),
 ('contraband_guard','Охранник контрабандистов','Защищает нелегальные грузы и тайные маршруты между складами. Не доверяет незнакомцам у охраняемых проходов.','humanoid'),
 ('gang_leader','Главарь банды','Руководит вооружённой группой в одном из закрытых районов. Считает местные переходы и убежища своей собственностью.','humanoid'),
 ('rogue_loader','Сбойный грузовой дроид','Промышленный дроид с повреждёнными ограничителями. Продолжает выполнять старые команды, не различая груз и препятствия.','droid'),
 ('rogue_sentry','Перепрограммированный дроид-страж','Охранный дроид с изменёнными списками доступа. Бывшие служебные маршруты стали его замкнутой зоной патрулирования.','droid'),
 ('ancient_automaton','Боевой автомат древнего сектора','Тяжёлая машина охраняет давно заброшенные узлы города. Её протоколы подчиняются приказам исчезнувших владельцев.','droid'),
 ('tunnel_beast','Подземный хищник','Обитает в тёмных пустотах между городскими ярусами. Движется по техническим шахтам и редко выходит к освещённым улицам.','beast'),
 ('jungle_predator','Джунглевый хищник','Скрывается под густыми кронами Дромунд-Кааса. Следит за добычей у троп и водопоев, оставаясь незаметным в зарослях.','beast'),
 ('swamp_predator','Болотный охотник','Терпеливо выжидает в затопленных низинах. Тёмная вода и переплетённые корни скрывают его приближение.','beast'),
 ('cave_predator','Пещерный зверь','Приспособился к жизни в ущельях и подземных проходах. Охраняет укрытия в глубине скал от любого вторжения.','beast'),
 ('alpha_predator','Матёрый хищник','Крупный зверь, занявший лучшие охотничьи угодья. Многочисленные шрамы напоминают о борьбе с соперниками.','beast'),
 ('imperial_deserter','Имперский дезертир','Скрывается вдали от гарнизонов и снабженческих постов. Сохранил военную подготовку и добывает припасы налётами.','humanoid'),
 ('tomb_raider','Расхититель гробниц','Проникает в древние сооружения в поисках ценных находок. Готов защищать тайные раскопки от случайных свидетелей.','humanoid'),
 ('secret_cultist','Адепт тайного культа','Оберегает забытые ритуальные места и оставленные святилища. В чужом присутствии видит угрозу своим тайнам.','humanoid'),
 ('guardian_droid','Древний дроид-хранитель','Механический страж сохранил часть древних охранных протоколов. Продолжает службу среди руин, лишённых прежних обитателей.','droid'),
 ('damaged_drill_droid','Повреждённый буровой дроид','Тяжёлая горная машина осталась без обслуживания в старых выработках. Повреждённые датчики заставляют её атаковать движущиеся цели.','droid'),
 ('tython_manka_cat','Манка-кот','Бесшумно движется среди тайтонских зарослей. Поджидает добычу у троп и водопоев, используя камни и высокую траву как укрытие.','beast'),
 ('tython_horranth','Хоррант','Крепкий зверь населяет склоны и лесные низины Тайтона. Защищая своё логово, способен преследовать нарушителя далеко за пределами укрытия.','beast'),
 ('tython_flesh_raider_scout','Разведчик плотоедов','Выслеживает незнакомцев у границ поселений своего племени. Пользуется зарослями и узкими проходами, чтобы незаметно приблизиться к цели.','humanoid'),
 ('tython_flesh_raider_hunter','Охотник плотоедов','Опытный налётчик знает скрытые тропы Тайтона. Выбирает места для засад и охраняет подходы к племенным стоянкам.','humanoid'),
 ('tython_flesh_raider_chief','Вожак плотоедов','Возглавляет вооружённый отряд у одного из племенных убежищ. Не терпит чужаков на охотничьих землях и собирает сородичей для нападения.','humanoid'),
 ('tython_war_droid','Тайтонский боевой дроид','Древняя боевая машина продолжает патрулировать руины Тайтона. Повреждённые системы опознавания считают любого посетителя угрозой.','droid'),
 ('tython_forge_guardian','Страж древней Кузницы','Тяжёлый механический хранитель охраняет проходы к Кузнице. Вековые повреждения не заглушили его защитные протоколы.','droid'),
 ('korriban_klorslug','К’лор-слизень','Покрытый прочным панцирем хищник обживает погребальные камеры Коррибана. Пробирается по трещинам в камне и нападает у входов в своё гнездо.','beast'),
 ('korriban_klorslug_broodmother','Матка к’лор-слизней','Крупная особь охраняет кладку в глубине гнезда. Её территория занимает несколько соединённых подземных камер.','beast'),
 ('korriban_shyrack','Ширак','Крылатый хищник с острыми клыками скрывается в пещерах Коррибана. Срывается со свода, услышав движение внизу.','beast'),
 ('korriban_tukata','Тук’ата','Свирепый зверь бродит среди скал и древних гробниц. Нападает на тех, кто приближается к его убежищу или отстаёт от экспедиции.','beast'),
 ('korriban_tukata_alpha','Вожак тук’ата','Матёрый хищник удерживает за своей стаей обширные угодья. Его появление выдаёт настороженное поведение остальных зверей.','beast'),
 ('korriban_renegade_acolyte','Послушник-отступник','Ищет силу в древних гробницах вне надзора наставников. Защищает найденные тайники и считает других искателей своими соперниками.','humanoid')
ON CONFLICT (code) DO NOTHING;

INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,
       LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('coruscant','lower_residential_levels','underworld_scavenger','gang_enforcer','gang_leader'),
 ('coruscant','service_walkways','rogue_loader','underworld_scavenger','rogue_sentry'),
 ('coruscant','freight_docks','underworld_scavenger','contraband_guard','gang_leader'),
 ('coruscant','workshop_quarter','rogue_loader','gang_enforcer','rogue_sentry'),
 ('coruscant','lower_markets','contraband_guard','hired_sharpshooter','gang_leader'),
 ('coruscant','abandoned_factories','rogue_loader','rogue_sentry','ancient_automaton'),
 ('coruscant','transit_tunnels','tunnel_beast','hired_sharpshooter','gang_leader'),
 ('coruscant','power_sublevels','rogue_sentry','gang_enforcer','ancient_automaton'),
 ('coruscant','sealed_sectors','tunnel_beast','rogue_sentry','ancient_automaton'),
 ('coruscant','deep_underlevels','tunnel_beast','gang_leader','ancient_automaton'),
 ('dromund_kaas','kaas_city_outskirts','jungle_predator','imperial_deserter','alpha_predator'),
 ('dromund_kaas','jungle_road','jungle_predator','tomb_raider','alpha_predator'),
 ('dromund_kaas','supply_outpost','imperial_deserter','guardian_droid','imperial_deserter'),
 ('dromund_kaas','quarry_trails','damaged_drill_droid','imperial_deserter','damaged_drill_droid'),
 ('dromund_kaas','storm_marshes','swamp_predator','jungle_predator','alpha_predator'),
 ('dromund_kaas','forgotten_ruins','tomb_raider','guardian_droid','secret_cultist'),
 ('dromund_kaas','shadow_gorges','cave_predator','imperial_deserter','alpha_predator'),
 ('dromund_kaas','ancient_catacombs','tomb_raider','guardian_droid','secret_cultist'),
 ('dromund_kaas','ritual_grounds','secret_cultist','guardian_droid','secret_cultist'),
 ('dromund_kaas','sealed_sanctum','guardian_droid','secret_cultist','guardian_droid'),
 ('tython','forest_trails','tython_manka_cat','tython_flesh_raider_scout','tython_horranth'),
 ('tython','misty_slopes','tython_horranth','tython_flesh_raider_hunter','tython_flesh_raider_chief'),
 ('tython','river_gorges','tython_manka_cat','tython_horranth','tython_flesh_raider_hunter'),
 ('tython','woodland_ravine','tython_flesh_raider_scout','tython_flesh_raider_hunter','tython_flesh_raider_chief'),
 ('tython','kaleth_approach','tython_flesh_raider_scout','tython_war_droid','tython_flesh_raider_chief'),
 ('tython','lower_kaleth','tython_war_droid','tython_flesh_raider_hunter','guardian_droid'),
 ('tython','upper_kaleth','tython_war_droid','tython_war_droid','guardian_droid'),
 ('tython','mountain_pass','tython_horranth','tython_flesh_raider_hunter','tython_flesh_raider_chief'),
 ('tython','forge_approach','tython_flesh_raider_hunter','tython_war_droid','tython_forge_guardian'),
 ('tython','ancient_forge','tython_war_droid','tython_forge_guardian','tython_forge_guardian'),
 ('korriban','valley_approach','korriban_klorslug','korriban_shyrack','korriban_klorslug_broodmother'),
 ('korriban','red_dunes','korriban_shyrack','korriban_klorslug','korriban_tukata_alpha'),
 ('korriban','rock_passes','korriban_tukata','korriban_tukata','korriban_tukata_alpha'),
 ('korriban','shyrack_caves','korriban_shyrack','korriban_tukata','korriban_shyrack'),
 ('korriban','ruined_crypts','tomb_raider','korriban_klorslug_broodmother','guardian_droid'),
 ('korriban','abandoned_excavation','tomb_raider','korriban_renegade_acolyte','secret_cultist'),
 ('korriban','ajunta_pall_tomb','korriban_klorslug','korriban_klorslug_broodmother','secret_cultist'),
 ('korriban','tulak_hord_tomb','korriban_renegade_acolyte','korriban_tukata','korriban_tukata_alpha'),
 ('korriban','marka_ragnos_tomb','tomb_raider','korriban_renegade_acolyte','guardian_droid'),
 ('korriban','naga_sadow_tomb','guardian_droid','secret_cultist','korriban_renegade_acolyte')
) AS v(planet_code,location_code,normal_template,strong_template,elite_template)
JOIN planets p ON p.code=v.planet_code
JOIN locations l ON l.planet_id=p.id AND l.code=v.location_code
CROSS JOIN LATERAL (VALUES
 ('normal',1,0,'normal',v.normal_template),
 ('strong',2,2,'strong',v.strong_template),
 ('elite',3,4,'elite',v.elite_template)
) AS r(code,position,level_offset,rank,template_code)
JOIN npc_templates t ON t.code=r.template_code
ON CONFLICT (location_id,code) DO NOTHING;

-- Seed slots by stable placement code, never by editable rank or runtime state.
INSERT INTO npc_spawns(location_npc_id,slot)
SELECT n.id,slots.slot FROM location_npcs n
JOIN locations l ON l.id=n.location_id
JOIN planets p ON p.id=l.planet_id
CROSS JOIN LATERAL generate_series(1,CASE n.code WHEN 'normal' THEN 2 ELSE 1 END) AS slots(slot)
WHERE p.code IN ('coruscant','dromund_kaas','tython','korriban') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
