-- Third additional protected-core batch. Core membership/control stay in seed07.
-- Early Great Galactic War; districts and enemy roles are authored adaptations.
-- Anaxes uses Legends geography, without later Clone Wars or canon destruction.
-- Alsakan's later conflicts and Rendili's late-Republic ship projects are not present.
-- Krayiss locations surround the buried library-temple; its inner gate stays sealed.
-- Ch'hodos has native desert wildlife, not the later Cold War radiation mutations.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('anaxes_supply_raider','Похититель флотских припасов','Перехватывает небольшие грузы на удалённых дорогах Анаксеса. Прячет добычу в пустующих хозяйственных корпусах и защищает тайники оружием.','humanoid'),
 ('anaxes_infiltrator','Диверсант складских окраин','Пользуется поддельными пропусками для проникновения в комплексы снабжения. Устраивает укрытия в старых усадьбах и охраняет украденные схемы.','humanoid'),
 ('anaxes_drill_droid','Сбойный учебный автомат','Остался на закрытой тренировочной площадке без надёжного управления. Повреждённые ограничения заставляют его атаковать любого, кто попадает в зону наблюдения.','droid'),
 ('anaxes_estate_sentry','Перепрограммированный усадебный страж','Списанный охранник загородного владения получил новые коды от преступников. Защищает захваченные здания и больше не признаёт законные допуски.','droid'),
 ('alsakan_terrace_gangster','Бандит городских террас','Контролирует малолюдные переходы между старыми кварталами Алсакана. Грабит перевозчиков и держит вооружённые посты у тайных складов.','humanoid'),
 ('alsakan_artifact_fencer','Вооружённый скупщик древностей','Торгует предметами из разграбленных алсаканских коллекций. Хранит добычу в закрытых залах и не терпит свидетелей незаконных сделок.','humanoid'),
 ('alsakan_maintenance_droid','Неисправный реставрационный дроид','Обслуживал каменные фасады и мозаики старых городских районов. Сломанные датчики заставляют его применять рабочие инструменты против движущихся помех.','droid'),
 ('alsakan_vault_sentry','Взломанный страж коллекции','Бывший охранник частного собрания подчиняется похитителям ценностей. Следит за узкими проходами к хранилищам и атакует всех посторонних.','droid'),
 ('rendili_parts_thief','Похититель корабельных узлов','Присваивает оборудование наземных подрядчиков Рендили. Вывозит детали через закрытые грузовые дворы и удерживает нелегальные склады.','humanoid'),
 ('rendili_illicit_engineer','Вооружённый подпольный инженер','Собирает корабельные системы из краденых узлов. Охраняет свою мастерскую и переделывает производственные машины для защиты преступных заказчиков.','humanoid'),
 ('rendili_assembly_droid','Повреждённый монтажный дроид','Старая машина продолжает работу на остановленной сборочной линии. Из-за сбоя принимает приближающихся существ за материал, который нужно закрепить и обработать.','droid'),
 ('rendili_test_automaton','Неуправляемый испытательный автомат','Проверял прочность оборудования в закрытом испытательном корпусе. Потерял связь с диспетчером и применяет силовые захваты вне разрешённого рабочего цикла.','droid'),
 ('dromund_fels_sand_stalker','Хищник сухих ложбин','Прячется среди каменных гряд Дромунд-Фелса. Подстерегает добычу у старых караванных путей и укрытых от солнца выемок.','beast'),
 ('dromund_fels_relic_raider','Расхититель древностей Дромунд-Фелса','Проникает в покинутые раскопы и вывозит находки без разрешения властей. Держит оружие наготове, защищая тайники и пути отхода.','humanoid'),
 ('dromund_fels_excavator','Сбойный археологический бурильщик','Освобождал древние сооружения от каменных завалов. Повреждённая система опознавания превратила его рабочие инструменты в угрозу для посетителей раскопа.','droid'),
 ('dromund_fels_renegade','Отступник сухих святилищ','Ищет запретные знания в древних сооружениях вне имперского надзора. Подчинил себе часть оставленной техники и охраняет собранные реликвии.','humanoid'),
 ('krayiss_relic_poacher','Расхититель раскопов Крайисса','Следит за экспедициями у древнего обелиска Крайисса II. Похищает находки и припасы, скрывая их в самовольно прорытых ходах.','humanoid'),
 ('krayiss_rogue_researcher','Отступник тайной экспедиции','Исследует внешние сооружения библиотеки-храма без разрешения наставников. Удерживает свои раскопы оружием и не допускает соперников к собранным материалам.','humanoid'),
 ('krayiss_excavation_droid','Повреждённый раскопочный автомат','Разбирал грунт и обломки у погребённых сооружений Крайисса II. После потери управления продолжает расширять проходы, не замечая оказавшихся рядом существ.','droid'),
 ('krayiss_security_droid','Взломанный экспедиционный страж','Охранный дроид исследовательского лагеря захвачен нелегальными искателями реликвий. Защищает их оборудование и перекрывает проходы к тайным складам.','droid'),
 ('chhodos_dune_hunter','Пустынный охотник Ч''ходоса','Держится у подножия красных скал и в узких ложбинах. Использует тень каменных выступов, чтобы незаметно приблизиться к добыче.','beast'),
 ('chhodos_carrion_beast','Скальный падальщик Ч''ходоса','Разыскивает останки животных среди сухих русел и каменных осыпей. Защищает кормовые участки и нападает на тех, кто проходит слишком близко.','beast'),
 ('chhodos_salvage_raider','Налётчик красных пустошей','Грабит удалённые стоянки Ч''ходоса и разбирает старые сооружения. Охраняет найденные запасы в укрытых от солнца каменных постройках.','humanoid'),
 ('chhodos_broken_sentinel','Повреждённый каменный часовой','Механический страж древнего комплекса покрыт пылью и обломками облицовки. Его повреждённые протоколы считают любого посетителя нарушителем.','droid')
ON CONFLICT (code) DO NOTHING;

-- Republic core: Anaxes, Alsakan and Rendili. All locations are on the surface.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('anaxes','fleet_supply_landing','Посадочная площадка снабжения','Грузовые дороги соединяют небольшой терминал с наземными складами. За пределами оживлённых проходов похитители высматривают оставленные партии припасов.',1,1),
 ('anaxes','sirpar_woodland_road','Лесная дорога Сирпара','Старая дорога петляет между деревьями у подножия холмов. В стороне от неё скрыты стоянки налётчиков и взломанные охранные машины.',2,6),
 ('anaxes','abandoned_training_ground','Покинутая учебная площадка','Списанные мишени стоят среди закрытых тренировочных полос. Неисправные автоматы продолжают упражнения, не распознавая команды остановки.',3,11),
 ('anaxes','hillside_service_tunnels','Служебные тоннели под холмами','Технические переходы связывают старые склады и хозяйственные корпуса. Диверсанты используют забытые выходы для тайной перевозки украденного оборудования.',4,16),
 ('anaxes','sirpar_estate_outskirts','Окраины усадеб Сирпара','Террасы холмов заняты загородными владениями и узкими служебными дорогами. В пустующих пристройках обосновались похитители припасов и взломщики охраны.',5,21),
 ('anaxes','decommissioned_vehicle_depot','Списанный парк техники','Остановленные транспортные платформы заполнили огороженные дворы. Преступники разбирают пригодные узлы, а сбойная учебная техника мешает проходу.',6,26),
 ('anaxes','old_signal_station','Старая сигнальная станция','Неиспользуемые антенны возвышаются над лесным склоном. Диверсанты восстановили часть аппаратуры и защищают перехваченные служебные записи.',7,31),
 ('anaxes','seized_officers_estate','Захваченная офицерская усадьба','Пустующая загородная резиденция скрыта за высокими стенами. Вооружённая группа заняла хозяйственные дворы и перепрограммировала оставленных стражей.',8,36),
 ('anaxes','closed_tactical_archive','Закрытый учебный архив','В старом учебном корпусе сохранились карты и схемы снабжения. Взломщики удерживают хранилище, используя неисправные автоматы для охраны проходов.',9,41),
 ('anaxes','hidden_supply_fort','Тайный укреплённый склад','Заброшенное укрепление в холмах превращено в базу похитителей флотских грузов. Его хозяева собрали здесь украденные припасы и переделанную охранную технику.',10,46),
 ('alsakan','lower_transit_apron','Нижняя транспортная площадка','Грузовой терминал выходит к многоярусным улицам старого города. На малолюдных служебных съездах промышляют банды перевозчиков краденого.',1,1),
 ('alsakan','terraced_backstreets','Переулки городских террас','Лестницы и подъёмники связывают тесные дворы на разных уровнях. Вооружённые банды перекрывают боковые проходы и скрывают здесь награбленное.',2,6),
 ('alsakan','mosaic_courtyards','Мозаичные дворы','Старинные узоры ещё различимы под пылью закрытых площадей. Похитители ценной отделки обходят неисправных реставрационных дроидов и охраняют свои тайники.',3,11),
 ('alsakan','old_market_passages','Проходы старого рынка','За заколоченными торговыми залами сохранилась сеть крытых галерей. Подпольные скупщики проводят здесь сделки, выставляя охрану у выходов.',4,16),
 ('alsakan','mountain_service_galleries','Служебные галереи горных кварталов','Городские постройки уступами поднимаются по каменным склонам. Забытые технические галереи служат путями перевозки похищенных древностей.',5,21),
 ('alsakan','sealed_collector_vaults','Закрытые хранилища коллекционеров','Массивные двери отделяют старые собрания от пустующих жилых корпусов. Взломщики заняли вскрытые помещения и подчинили себе часть механической охраны.',6,26),
 ('alsakan','abandoned_civic_halls','Покинутые городские залы','Величественные помещения давно выведены из общего пользования. Незаконные разборщики вывозят уцелевшую отделку, защищая грузовые выходы оружием.',7,31),
 ('alsakan','antiquities_exchange','Подпольный рынок древностей','За неприметными складскими воротами выставлены похищенные предметы старых коллекций. Сделки проходят под охраной бандитов и взломанных стражей.',8,36),
 ('alsakan','old_family_residence','Старая родовая резиденция','Замкнутые дворы окружены пустующими корпусами древнего владения. Торговцы краденым превратили комнаты в тайные склады и восстановили охранные цепи.',9,41),
 ('alsakan','hidden_collection_keep','Укреплённое тайное хранилище','За несколькими рядами ворот собраны ценности из разграбленных алсаканских собраний. Вооружённые хозяева контролируют переходы вместе с перепрограммированными дроидами.',10,46),
 ('rendili','ground_supply_terminal','Наземный терминал подрядчиков','Грузовые площадки принимают детали для кораблестроительных предприятий Рендили. У дальних навесов похитители вскрывают контейнеры и подменяют ведомости.',1,1),
 ('rendili','component_yards','Дворы корабельных компонентов','Ряды наземных складов разделены погрузочными проездами. В закрытых секциях скрываются похитители деталей и неисправные монтажные дроиды.',2,6),
 ('rendili','old_assembly_bays','Старые монтажные ангары','Остановленные линии окружены забытыми стойками корабельного оборудования. Часть машин продолжает опасную работу, мешая нелегальным разборщикам.',3,11),
 ('rendili','factory_service_tunnels','Заводские служебные тоннели','Крытые технические переходы соединяют пустующие производственные корпуса. Подпольные инженеры проложили через них пути доставки краденых узлов.',4,16),
 ('rendili','decommissioned_motor_labs','Закрытые двигательные лаборатории','За усиленными перегородками сохранились стенды проверки силовых установок. Незаконные мастера запустили часть оборудования без штатных ограничений безопасности.',5,21),
 ('rendili','abandoned_vehicle_depot','Покинутая база наземной техники','Старые гаражи и ремонтные ямы занимают огороженный участок за заводами. Налётчики собирают здесь транспорт из найденных деталей и охраняют свои запасы.',6,26),
 ('rendili','design_storage','Хранилище инженерных схем','Закрытые архивные комнаты содержат старую производственную документацию. Похитители чертежей удерживают доступ к записям и оборудованию для их расшифровки.',7,31),
 ('rendili','unauthorized_refit_halls','Подпольные цеха переоборудования','В пустующих ангарах собраны нелегальные линии ремонта корабельных систем. Вооружённые инженеры защищают свои заказы с помощью переделанной автоматики.',8,36),
 ('rendili','heavy_test_enclosure','Закрытый испытательный корпус','Массивные опоры окружают площадки проверки тяжёлых механизмов. Потерявшие управление автоматы продолжают силовые испытания в пустых залах.',9,41),
 ('rendili','seized_engineering_complex','Захваченный инженерный комплекс','Группа похитителей заняла наземный комплекс одного из подрядчиков верфей. Здесь собраны краденые разработки, производственные машины и техника нелегальных мастерских.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Imperial core: Dromund Fels, Krayiss II and Ch'hodos.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('dromund_fels','arid_supply_landing','Засушливая площадка снабжения','Небольшой терминал стоит у дороги к удалённым исследовательским участкам. За дальними грузовыми навесами скрываются похитители экспедиционных припасов.',1,1),
 ('dromund_fels','dry_caravan_tracks','Сухие караванные пути','Колеи старых платформ пересекают каменистые равнины. Хищники подстерегают путников в низинах, а налётчики следят за грузами с укрытых стоянок.',2,6),
 ('dromund_fels','stone_hollows','Каменные ложбины','Узкие впадины дают укрытие от сухого ветра и солнца. Между звериными логовами встречаются тайники незаконных искателей древностей.',3,11),
 ('dromund_fels','abandoned_dig_camp','Покинутый раскопочный лагерь','Пустые модули окружают старые выемки у каменных стен. Расхитители разбирают оставленную технику и охраняют найденные предметы.',4,16),
 ('dromund_fels','fractured_excavation_steps','Разрушенные уступы раскопа','Широкие ступени спускаются к основаниям древнего комплекса. Сбойные бурильщики продолжают разбирать породу рядом с нелегальными экспедициями.',5,21),
 ('dromund_fels','smugglers_caravan_shelter','Караванное укрытие контрабандистов','Грузовые навесы спрятаны за высокой каменной грядой. Здесь собирают находки из самовольных раскопок и готовят их к тайному вывозу.',6,26),
 ('dromund_fels','buried_outer_chambers','Засыпанные внешние палаты','Из-под обломков открылись служебные залы старого сооружения. Отступники удерживают проходы и используют неисправную технику для новых раскопов.',7,31),
 ('dromund_fels','sealed_relic_shelter','Закрытое хранилище находок','Восстановленные двери скрывают временные склады нелегальной экспедиции. Вооружённые исследователи охраняют собранные реликвии и записи о древних сооружениях.',8,36),
 ('dromund_fels','old_sanctuary_courts','Дворы старого святилища','Высокие каменные стены окружают расчищенные площадки древнего комплекса. Здесь закрепились отступники, подчинившие себе оставленные раскопочные машины.',9,41),
 ('dromund_fels','hidden_relic_enclave','Тайный оплот искателей реликвий','Удалённые залы соединены укреплёнными переходами и закрытыми грузовыми дворами. Хозяева убежища хранят здесь запретные находки и держат вооружённые посты.',10,46),
 ('krayiss_ii','survey_landing','Посадочная площадка исследователей','Расчищенные опоры окружают небольшой экспедиционный терминал. На удалённых складах похитители разыскивают оборудование для собственных раскопок.',1,1),
 ('krayiss_ii','old_expedition_camps','Старые экспедиционные стоянки','Оставленные лагеря отмечают пути прежних исследовательских групп. Расхитители занимают уцелевшие модули и восстанавливают механическую охрану.',2,6),
 ('krayiss_ii','obelisk_approach','Подступы к обелиску','Над местом погребённой библиотеки-храма возвышается древний каменный обелиск. В стороне от основного подхода нелегальные искатели устроили скрытые наблюдательные посты.',3,11),
 ('krayiss_ii','collapsed_supply_trench','Обрушенная траншея снабжения','Старый технический проход частично засыпан грунтом и обломками. Через уцелевшие участки расхитители переносят припасы, избегая сбойных раскопочных машин.',4,16),
 ('krayiss_ii','stone_fragment_yard','Площадка каменных фрагментов','За низкими ограждениями собраны обломки из внешних раскопов. Отступники охраняют сортировочные столы и технику для изучения найденных надписей.',5,21),
 ('krayiss_ii','sealed_survey_station','Закрытая исследовательская станция','Пустующие лабораторные модули окружают укреплённый архив экспедиции. Незаконные хозяева восстановили часть аппаратуры и взломали охранных дроидов.',6,26),
 ('krayiss_ii','illicit_dig_tunnels','Самовольные раскопочные ходы','Узкие тоннели проложены между внешними выемками вокруг погребённого комплекса. Их хозяева защищают свои находки и продолжают опасную работу неисправными машинами.',7,31),
 ('krayiss_ii','relic_sorting_vault','Тайное хранилище раскопщиков','Укреплённые контейнеры и переносные стеллажи заполняют скрытый служебный отсек. Здесь собирают похищенные образцы, документы и снаряжение исследователей.',8,36),
 ('krayiss_ii','obelisk_excavation_rim','Край раскопа у обелиска','Подпорные конструкции удерживают грунт у внешней границы древнего сооружения. Отступники выставили стражей у рабочих площадок и складов найденных фрагментов.',9,41),
 ('krayiss_ii','library_outer_excavation','Внешний раскоп библиотеки-храма','Расчищенные выемки проходят вдоль погребённых внешних стен. Внутренние врата остаются закрыты, а незаконная экспедиция удерживает доступные площадки и свои склады.',10,46),
 ('chhodos','red_rock_landing','Посадочная площадка красных скал','Каменные уступы окружают небольшой грузовой терминал под красным солнцем. На дальних подступах скрываются налётчики и пустынные звери.',1,1),
 ('chhodos','dry_riverbeds','Сухие каменные русла','Разветвлённые ложбины пересекают красные пустоши. Хищники укрываются в тени стен, а падальщики разыскивают останки животных среди осыпей.',2,6),
 ('chhodos','desert_shelters','Пустынные укрытия','Каменные навесы и старые стены защищают заброшенные стоянки. Налётчики прячут здесь припасы, отгоняя зверей от своих складов.',3,11),
 ('chhodos','yellow_stone_outworks','Укрепления жёлтого камня','Древние каменные постройки возвышаются над скалистым участком. Расхитители заняли уцелевшие дворы и обходят повреждённых механических часовых.',4,16),
 ('chhodos','scavenger_caverns','Пещеры падальщиков','Природные полости уходят под нагретые солнцем скалы. На полу лежат кости пустынных животных, а в глубине скрываются их опасные соседи.',5,21),
 ('chhodos','ancient_muster_ground','Древняя сборная площадь','Широкая площадка сохранилась перед стенами старого военного комплекса. Вооружённые разборщики охраняют найденные запасы и пути к заброшенным корпусам.',6,26),
 ('chhodos','ruined_storehouse_rows','Разрушенные складские ряды','Обвалившиеся крыши открывают проходы между древними кладовыми. Среди каменных завалов бродят падальщики, а у вскрытых хранилищ стоят посты налётчиков.',7,31),
 ('chhodos','seized_stone_redoubt','Захваченный каменный редут','Высокие стены скрывают площадки нелегальных перевозчиков. Хозяева укрепления восстановили старых часовых и заставили их охранять награбленное.',8,36),
 ('chhodos','red_cliff_passages','Проходы красных обрывов','Узкие тропы связывают древние сооружения над сухими низинами. Хищники держатся у естественных ниш, а налётчики перекрывают удобные переходы.',9,41),
 ('chhodos','sealed_ancient_stronghold','Закрытое древнее укрепление','Внутренние дворы старого комплекса окружены толстыми стенами жёлтого камня. Расхитители удерживают уцелевшие помещения вместе с повреждённой механической охраной.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Three placements per location; two normal slots and one each of strong/elite.
INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('anaxes','fleet_supply_landing','anaxes_supply_raider','rogue_loader','anaxes_infiltrator'),
 ('anaxes','sirpar_woodland_road','anaxes_supply_raider','anaxes_estate_sentry','anaxes_supply_raider'),
 ('anaxes','abandoned_training_ground','anaxes_drill_droid','anaxes_supply_raider','anaxes_drill_droid'),
 ('anaxes','hillside_service_tunnels','anaxes_infiltrator','anaxes_estate_sentry','anaxes_infiltrator'),
 ('anaxes','sirpar_estate_outskirts','anaxes_supply_raider','anaxes_infiltrator','anaxes_estate_sentry'),
 ('anaxes','decommissioned_vehicle_depot','anaxes_supply_raider','anaxes_drill_droid','anaxes_drill_droid'),
 ('anaxes','old_signal_station','anaxes_infiltrator','anaxes_estate_sentry','anaxes_infiltrator'),
 ('anaxes','seized_officers_estate','anaxes_infiltrator','anaxes_estate_sentry','anaxes_estate_sentry'),
 ('anaxes','closed_tactical_archive','anaxes_infiltrator','anaxes_drill_droid','anaxes_estate_sentry'),
 ('anaxes','hidden_supply_fort','anaxes_supply_raider','anaxes_infiltrator','anaxes_estate_sentry'),
 ('alsakan','lower_transit_apron','alsakan_terrace_gangster','rogue_loader','alsakan_terrace_gangster'),
 ('alsakan','terraced_backstreets','alsakan_terrace_gangster','alsakan_maintenance_droid','alsakan_artifact_fencer'),
 ('alsakan','mosaic_courtyards','alsakan_terrace_gangster','alsakan_maintenance_droid','alsakan_maintenance_droid'),
 ('alsakan','old_market_passages','alsakan_terrace_gangster','alsakan_artifact_fencer','alsakan_vault_sentry'),
 ('alsakan','mountain_service_galleries','alsakan_terrace_gangster','alsakan_maintenance_droid','alsakan_artifact_fencer'),
 ('alsakan','sealed_collector_vaults','alsakan_artifact_fencer','alsakan_vault_sentry','alsakan_vault_sentry'),
 ('alsakan','abandoned_civic_halls','alsakan_terrace_gangster','alsakan_maintenance_droid','alsakan_artifact_fencer'),
 ('alsakan','antiquities_exchange','alsakan_artifact_fencer','alsakan_terrace_gangster','alsakan_vault_sentry'),
 ('alsakan','old_family_residence','alsakan_artifact_fencer','alsakan_vault_sentry','alsakan_vault_sentry'),
 ('alsakan','hidden_collection_keep','alsakan_terrace_gangster','alsakan_artifact_fencer','alsakan_vault_sentry'),
 ('rendili','ground_supply_terminal','rendili_parts_thief','rogue_loader','rendili_illicit_engineer'),
 ('rendili','component_yards','rendili_parts_thief','rendili_assembly_droid','rendili_parts_thief'),
 ('rendili','old_assembly_bays','rendili_assembly_droid','rendili_parts_thief','rendili_assembly_droid'),
 ('rendili','factory_service_tunnels','rendili_parts_thief','rendili_assembly_droid','rendili_illicit_engineer'),
 ('rendili','decommissioned_motor_labs','rendili_illicit_engineer','rendili_assembly_droid','rendili_test_automaton'),
 ('rendili','abandoned_vehicle_depot','rendili_parts_thief','rendili_illicit_engineer','rendili_assembly_droid'),
 ('rendili','design_storage','rendili_illicit_engineer','rogue_sentry','rendili_illicit_engineer'),
 ('rendili','unauthorized_refit_halls','rendili_illicit_engineer','rendili_assembly_droid','rendili_test_automaton'),
 ('rendili','heavy_test_enclosure','rendili_test_automaton','rendili_assembly_droid','rendili_test_automaton'),
 ('rendili','seized_engineering_complex','rendili_parts_thief','rendili_illicit_engineer','rendili_test_automaton'),
 ('dromund_fels','arid_supply_landing','dromund_fels_relic_raider','dromund_fels_sand_stalker','dromund_fels_excavator'),
 ('dromund_fels','dry_caravan_tracks','dromund_fels_sand_stalker','dromund_fels_relic_raider','dromund_fels_sand_stalker'),
 ('dromund_fels','stone_hollows','dromund_fels_sand_stalker','dromund_fels_relic_raider','dromund_fels_relic_raider'),
 ('dromund_fels','abandoned_dig_camp','dromund_fels_relic_raider','dromund_fels_excavator','dromund_fels_relic_raider'),
 ('dromund_fels','fractured_excavation_steps','dromund_fels_relic_raider','dromund_fels_excavator','dromund_fels_excavator'),
 ('dromund_fels','smugglers_caravan_shelter','dromund_fels_relic_raider','contraband_guard','dromund_fels_renegade'),
 ('dromund_fels','buried_outer_chambers','dromund_fels_renegade','dromund_fels_excavator','dromund_fels_renegade'),
 ('dromund_fels','sealed_relic_shelter','dromund_fels_relic_raider','dromund_fels_excavator','dromund_fels_renegade'),
 ('dromund_fels','old_sanctuary_courts','dromund_fels_renegade','dromund_fels_excavator','dromund_fels_renegade'),
 ('dromund_fels','hidden_relic_enclave','dromund_fels_relic_raider','dromund_fels_excavator','dromund_fels_renegade'),
 ('krayiss_ii','survey_landing','krayiss_relic_poacher','rogue_loader','krayiss_security_droid'),
 ('krayiss_ii','old_expedition_camps','krayiss_relic_poacher','krayiss_security_droid','krayiss_relic_poacher'),
 ('krayiss_ii','obelisk_approach','krayiss_relic_poacher','krayiss_security_droid','krayiss_rogue_researcher'),
 ('krayiss_ii','collapsed_supply_trench','krayiss_relic_poacher','krayiss_excavation_droid','krayiss_excavation_droid'),
 ('krayiss_ii','stone_fragment_yard','krayiss_rogue_researcher','krayiss_excavation_droid','krayiss_security_droid'),
 ('krayiss_ii','sealed_survey_station','krayiss_rogue_researcher','krayiss_security_droid','krayiss_rogue_researcher'),
 ('krayiss_ii','illicit_dig_tunnels','krayiss_relic_poacher','krayiss_excavation_droid','krayiss_rogue_researcher'),
 ('krayiss_ii','relic_sorting_vault','krayiss_relic_poacher','krayiss_security_droid','krayiss_rogue_researcher'),
 ('krayiss_ii','obelisk_excavation_rim','krayiss_rogue_researcher','krayiss_excavation_droid','krayiss_security_droid'),
 ('krayiss_ii','library_outer_excavation','krayiss_relic_poacher','krayiss_security_droid','krayiss_rogue_researcher'),
 ('chhodos','red_rock_landing','chhodos_salvage_raider','chhodos_dune_hunter','chhodos_salvage_raider'),
 ('chhodos','dry_riverbeds','chhodos_carrion_beast','chhodos_dune_hunter','chhodos_dune_hunter'),
 ('chhodos','desert_shelters','chhodos_salvage_raider','chhodos_carrion_beast','chhodos_dune_hunter'),
 ('chhodos','yellow_stone_outworks','chhodos_salvage_raider','chhodos_broken_sentinel','chhodos_broken_sentinel'),
 ('chhodos','scavenger_caverns','chhodos_carrion_beast','chhodos_dune_hunter','chhodos_carrion_beast'),
 ('chhodos','ancient_muster_ground','chhodos_salvage_raider','chhodos_broken_sentinel','chhodos_salvage_raider'),
 ('chhodos','ruined_storehouse_rows','chhodos_carrion_beast','chhodos_salvage_raider','chhodos_broken_sentinel'),
 ('chhodos','seized_stone_redoubt','chhodos_salvage_raider','chhodos_broken_sentinel','chhodos_broken_sentinel'),
 ('chhodos','red_cliff_passages','chhodos_dune_hunter','chhodos_carrion_beast','chhodos_salvage_raider'),
 ('chhodos','sealed_ancient_stronghold','chhodos_salvage_raider','chhodos_broken_sentinel','chhodos_broken_sentinel')
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

INSERT INTO npc_spawns(location_npc_id,slot)
SELECT n.id,slots.slot FROM location_npcs n
JOIN locations l ON l.id=n.location_id
JOIN planets p ON p.id=l.planet_id
CROSS JOIN LATERAL generate_series(1,CASE n.code WHEN 'normal' THEN 2 ELSE 1 END) AS slots(slot)
WHERE p.code IN ('anaxes','alsakan','rendili','dromund_fels','krayiss_ii','chhodos') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
