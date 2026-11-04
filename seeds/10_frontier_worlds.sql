-- First frontier batch. Ownership is defined in seed07 and is never reset here.
-- Biomes and ancient sites follow Legends/SWTOR references; areas and enemy roles
-- are authored for the early Great Galactic War, before the Treaty of Coruscant.
-- No later Ord Mantell civil war, occupied Balmorra storyline, Ossus Jedi colony,
-- or named Athiss flashpoint expedition/bosses are imported.

-- Republic holdings: Arkania, Balmorra and Ord Mantell.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('arkania','snowbound_landing','Заснеженная посадочная площадка','Ветер заметает снегом огни небольшого посадочного узла. За рядами грузовых контейнеров начинается дорога к шахтёрским поселениям.',1,1),
 ('arkania','tundra_trails','Тропы тундры','Колеи вездеходов пересекают белую равнину между низкими скалами. В складках рельефа скрываются звериные логова и временные укрытия охотников.',2,6),
 ('arkania','ice_gorges','Ледяные ущелья','Глубокие трещины разрезают промёрзшее плато. Снежные мосты над провалами чередуются с узкими уступами, где трудно разминуться со встречным путником.',3,11),
 ('arkania','mining_camp','Старый шахтёрский лагерь','Обледенелые жилые модули окружают вход в давно закрытую выработку. Внутри ещё гудят автономные генераторы, привлекая тех, кто ищет тепло и забытые припасы.',4,16),
 ('arkania','diamond_mines','Алмазные шахты','Под слоями льда лежат богатые минеральные жилы. Свет буровых установок выхватывает из темноты крепления, грузовые рельсы и блеск неразработанной породы.',5,21),
 ('arkania','mining_tunnels','Глубокие штольни','Нижние выработки давно лишились постоянного обслуживания. За завалами сохранились проходы к старым подъёмникам и машинам, продолжающим выполнять последние команды.',6,26),
 ('arkania','research_outskirts','Окраины исследовательского комплекса','Закрытые научные корпуса соединены крытыми переходами. На дальних площадках хранятся списанное оборудование и контейнеры с неполной маркировкой.',7,31),
 ('arkania','sealed_laboratories','Закрытые лаборатории','Несколько исследовательских секций отрезаны тяжёлыми шлюзами. В пустых коридорах мерцают аварийные панели, а охранные системы всё ещё откликаются на движение.',8,36),
 ('arkania','glacier_caverns','Пещеры под ледником','Талая вода выточила под ледяной толщей просторные галереи. Отполированные стены усиливают каждый звук, и следы крупных существ ведут всё глубже.',9,41),
 ('arkania','dragon_cliffs','Драконьи скалы','Острые вершины выступают над морем снега. На защищённых от ветра уступах лежат крупные гнёзда, а под ними тянутся следы немногих вернувшихся экспедиций.',10,46),
 ('balmorra','supply_platforms','Снабженческие платформы','Транспортные площадки принимают детали и сырьё для заводских районов. Между контейнерами проходят служебные маршруты, которыми пользуются и незваные посетители.',1,1),
 ('balmorra','workshop_streets','Улицы мастерских','Вдоль узких улиц стоят ремонтные ангары и небольшие сборочные цеха. Из открытых ворот доносится стук станков, а переулки забиты списанными механизмами.',2,6),
 ('balmorra','scrap_yards','Заводские свалки','Горы металлических обрезков разделены ржавыми ограждениями. Среди разбитых корпусов проложены тайные тропы сборщиков лома и похитителей оборудования.',3,11),
 ('balmorra','assembly_district','Сборочный квартал','Многоэтажные производственные линии соединяют конвейеры и подвесные мостки. Часть цехов работает без людей, подчиняясь уцелевшим автоматическим диспетчерам.',4,16),
 ('balmorra','freight_tunnels','Грузовые тоннели','Под промышленными кварталами идут широкие транспортные галереи. В боковых тупиках скопились забытые грузы, и сигналы поездов слышны задолго до их появления.',5,21),
 ('balmorra','chemical_wastes','Промышленные пустоши','За очистными сооружениями лежит полоса изрытой земли. Старые отстойники и редкие заросли дают укрытие хищникам, привыкшим к соседству заводов.',6,26),
 ('balmorra','abandoned_foundry','Заброшенный литейный цех','Над остывшими печами нависают грузовые краны. Между литейными формами сохранились запасы металла и тяжёлые машины с повреждёнными системами управления.',7,31),
 ('balmorra','weapons_warehouses','Оружейные склады','Бронированные хранилища отделены друг от друга контрольными проходами. Сорванные пломбы и свежие следы погрузки выдают нелегальные сделки с заводским вооружением.',8,36),
 ('balmorra','testing_grounds','Испытательные полигоны','За высокими насыпями тянутся полосы препятствий и укреплённые огневые позиции. Автоматические мишени соседствуют с машинами, так и не завершившими испытательный цикл.',9,41),
 ('balmorra','prototype_vault','Хранилище прототипов','В изолированных боксах остались опытные боевые системы. Резервное питание поддерживает замки и охрану, а некоторые стенды продолжают передавать команды своим машинам.',10,46),
 ('ord_mantell','cargo_quay','Грузовой причал','У потемневших посадочных платформ разгружают корабли с окраинных маршрутов. Под навесами стоят ящики без документов, за которыми внимательно следят вооружённые люди.',1,1),
 ('ord_mantell','coastal_tracks','Прибрежные тропы','Тропы между каменистыми холмами спускаются к узким пляжам. Вдали от причалов рыбацкие стоянки соседствуют с неприметными перевалочными точками контрабандистов.',2,6),
 ('ord_mantell','junk_market','Рынок старых деталей','Ряды прилавков устроены прямо среди разобранных машин. Здесь можно найти редкие узлы, но споры о праве собственности часто решают местные банды.',3,11),
 ('ord_mantell','wreck_fields','Поля обломков','Полузасыпанные корпуса кораблей лежат между холмами. В их отсеках укрываются налётчики, а уцелевшие механизмы мешают безопасно разбирать старый металл.',4,16),
 ('ord_mantell','scavenger_camps','Лагеря налётчиков','За стенами из корабельных панелей спрятаны мастерские и склады награбленного. Узкие проходы между лагерями позволяют вооружённым группам быстро исчезать среди обломков.',5,21),
 ('ord_mantell','lava_ravines','Лавовые овраги','Древние потоки лавы оставили тёмные гряды и глубокие расселины. Между ними проходят трудные обходные маршруты, удобные для засад и скрытых стоянок.',6,26),
 ('ord_mantell','smugglers_coves','Бухты контрабандистов','Скалистый берег скрывает несколько защищённых бухт. Небольшие суда подходят к тайникам под прикрытием утёсов, минуя официальные грузовые терминалы.',7,31),
 ('ord_mantell','pirate_docks','Пиратские доки','В глубине бухты оборудованы ремонтные площадки для захваченных судов. Между ангарами выставлены посты, а грузовые переходы перегорожены импровизированной бронёй.',8,36),
 ('ord_mantell','savrip_highlands','Нагорья саврипов','Высокие каменистые уступы отделяют земли саврипов от людных дорог. Среди скал видны укрытия и грубые заграждения, предупреждающие о чужой территории.',9,41),
 ('ord_mantell','raider_stronghold','Крепость налётчиков','Укреплённое убежище занимает остов крупного корабля у подножия скал. Внутри устроены оружейные комнаты, наблюдательные посты и защищённые склады добычи.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Imperial holdings: Athiss, Thule and Ossus.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('athiss','expedition_landing','Старая площадка экспедиции','Небольшой участок твёрдой земли сохранил следы посадочных опор. Заброшенные контейнеры почти скрылись в зелени, и от площадки осталась лишь тропа к древним постройкам.',1,1),
 ('athiss','overgrown_trail','Заросшая тропа','Крупные корни разбили старую каменную дорогу. Под густыми кронами трудно увидеть поворот заранее, а движение в зарослях редко принадлежит мирному путнику.',2,6),
 ('athiss','vine_gorge','Ущелье лиан','Скальные стены увиты плотными плетями растений. По дну течёт мелкий поток, над которым нависают остатки мостов и балконов забытых сооружений.',3,11),
 ('athiss','sunken_ruins','Затонувшие руины','Низкие залы древнего комплекса заполнила тёмная вода. Над поверхностью выступают каменные плиты, а между затопленными стенами проложены ненадёжные переходы.',4,16),
 ('athiss','broken_colonnade','Разрушенная колоннада','Ряды расколотых колонн отмечают вход в храмовую область. Растения обвили древние барельефы, но проходы между ними выглядят недавно расчищенными.',5,21),
 ('athiss','temple_courtyard','Храмовый двор','За внешними стенами открывается широкая площадка с остатками статуй. На истёртых плитах заметны следы повторяющихся обрядов и тяжёлых звериных лап.',6,26),
 ('athiss','ritual_halls','Ритуальные залы','Каменные помещения соединены узкими коридорами. Старые символы на стенах обведены заново, а в нишах хранятся предметы неизвестного назначения.',7,31),
 ('athiss','alchemical_chambers','Алхимические покои','В глубине комплекса сохранились чаши, клетки и массивные рабочие столы. Из закрытых отсеков слышится движение, не похожее на шаги храмовых служителей.',8,36),
 ('athiss','beast_pits','Звериные ямы','Подземные заграждения отделяют узкие галереи от глубоких каменных загонов. Некоторые решётки выломаны изнутри, и следы когтей покрывают даже высокие стены.',9,41),
 ('athiss','buried_sanctum','Погребённое святилище','Самые старые помещения скрыты под поздними пристройками храма. Плотные двери отделяют их от внешних залов, а служители культа тщательно охраняют ведущие вниз лестницы.',10,46),
 ('thule','dry_steppe','Сухие степи','Жёсткая трава растёт между полосами растрескавшейся земли. Над равниной собираются тёмные облака, но вспышки молний редко сопровождаются долгожданным дождём.',1,1),
 ('thule','storm_pass','Грозовой перевал','Каменная тропа пересекает открытый всем ветрам хребет. Грозовые разряды освещают дальние склоны, а немногочисленные укрытия давно заняты незваными обитателями.',2,6),
 ('thule','basalt_gorges','Базальтовые ущелья','Чёрные стены разломов спускаются к высохшим руслам. В тесных проходах ветер поднимает пыль, скрывая следы зверей и небольших вооружённых отрядов.',3,11),
 ('thule','abandoned_waystation','Заброшенная дорожная станция','Старый дорожный узел стоит на пересечении степных маршрутов. В уцелевших ангарах находят укрытие мародёры, а служебные машины охраняют давно пустые площадки.',4,16),
 ('thule','ruin_fields','Поля древних руин','На равнине разбросаны основания стен и каменные обломки. Между ними встречаются спуски в погребённые помещения и свежие следы искателей древних ценностей.',5,21),
 ('thule','cracked_aqueducts','Расколотые акведуки','Сломанные арки древней системы водоснабжения пересекают сухие низины. Под уцелевшими перекрытиями устроены тайники и переходы к подземным резервуарам.',6,26),
 ('thule','underground_cisterns','Подземные цистерны','Массивные каменные камеры когда-то собирали воду для поселений. Теперь в их глубине скрываются сторожевые механизмы и группы, избегающие открытых дорог.',7,31),
 ('thule','sith_bastion','Древний бастион ситхов','Укрепление возвышается над степью, сохраняя остатки старых оборонительных линий. Внутренние дворы соединены узкими воротами, удобными для небольшой охраны.',8,36),
 ('thule','lightning_spires','Шпили молний','Высокие сооружения древнего комплекса выступают над грозовыми облаками. На площадках между ними лежат истёртые ритуальные знаки и обломки защитных машин.',9,41),
 ('thule','sealed_war_vault','Запечатанный боевой арсенал','Под основанием бастиона расположены тяжёлые хранилища старой военной техники. Вскрытые печати открыли проход к машинам, которые так и не получили приказа завершить службу.',10,46),
 ('ossus','dusty_landing','Пыльная посадочная площадка','Ровная площадка расчищена среди обломков древних построек. За сигнальными огнями начинается пустошь, где старые дороги почти исчезли под слоями пыли.',1,1),
 ('ossus','ash_plains','Пепельные равнины','Ветер несёт пыль над потрескавшейся землёй. Редкие укрытия у скал позволяют переждать бурю, но их приходится делить с местными хищниками.',2,6),
 ('ossus','wreckage_hollows','Лощины обломков','В складках рельефа лежат остатки древних кораблей и разбитых машин. Уцелевшие отсеки привлекают искателей запчастей, готовых защищать найденное оружием.',3,11),
 ('ossus','ruined_avenues','Разрушенные проспекты','Широкие улицы исчезнувшего города обрываются у завалов. Между основаниями зданий проходят тропы мародёров и редкие маршруты сохранившихся дроидов.',4,16),
 ('ossus','rock_shelters','Скальные убежища','У подножия гор сохранились пещеры и высеченные в камне помещения. Их сухие залы служат стоянками охотникам за реликвиями и логовами зверей.',5,21),
 ('ossus','buried_archives','Погребённые архивы','Под обрушившимися перекрытиями лежат остатки древних хранилищ знаний. В узких проходах встречаются повреждённые терминалы и запертые отсеки, которых ещё не коснулись раскопки.',6,26),
 ('ossus','library_approach','Подступы к Великой библиотеке','Осколки монументов и старые лестницы ведут к разрушенному библиотечному комплексу. Масштаб уцелевших оснований напоминает о месте, которое когда-то занимал Оссус в жизни джедаев.',7,31),
 ('ossus','ruined_galleries','Разрушенные галереи','Длинные залы частично открыты небу. Между поваленными опорами сохранились защищённые ниши, а в боковых проходах продолжают движение древние механические стражи.',8,36),
 ('ossus','sealed_stacks','Запечатанные книгохранилища','За тяжёлыми дверями уцелели отдельные секции древнего собрания. Искатели реликвий оставили здесь инструменты и пробные вскрытия, но многие проходы всё ещё охраняются.',9,41),
 ('ossus','deep_repository','Глубинное хранилище','Нижние камеры скрыты под несколькими слоями разрушенных залов. В тишине слышна работа резервных механизмов, защищающих то, что осталось от древнего собрания.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('arkania_tundra_stalker','Тундровый хищник','Прячется среди снежных наносов и невысоких скал Аркании. Подолгу идёт по следу добычи, ожидая удобного места для нападения.','beast'),
 ('arkania_corporate_mercenary','Корпоративный наёмник','Охраняет закрытые разработки и незаконные грузы частных заказчиков. Считает чужое присутствие у секретных объектов достаточным поводом открыть огонь.','humanoid'),
 ('arkania_lab_droid','Лабораторный дроид-охранник','Следит за изолированными секциями исследовательского комплекса. После утраты связи с диспетчером воспринимает любого посетителя как нарушителя карантина.','droid'),
 ('arkania_dragon','Дракон Аркании','Крылатый обитатель ледяных гор охраняет удалённые места гнездования. Внимательно наблюдает за приближающимися путниками и запоминает пути к их стоянкам.','beast'),
 ('balmorra_factory_saboteur','Заводской диверсант','Проникает в цеха под видом ремонтного рабочего. Уничтожает оборудование, похищает чертежи и устраняет тех, кто обнаружил его тайники.','humanoid'),
 ('balmorra_security_droid','Заводской охранный дроид','Перепрограммирован для охраны захваченных складов и закрытых производственных линий. Проверяет посетителей по изменённому списку допуска.','droid'),
 ('balmorra_assault_prototype','Прототип штурмового дроида','Опытная боевая машина осталась под управлением испытательной программы. Продолжает искать цели даже за пределами отведённого ей полигона.','droid'),
 ('balmorra_wasteland_scavenger','Падальщик промышленных пустошей','Выживает среди отстойников и заброшенных заводских площадок. Использует трубы и заросли, чтобы подобраться к ослабевшей добыче.','beast'),
 ('ord_mantell_scrap_raider','Налётчик со свалок','Добывает детали грабежом и силой отбирает ценные находки. Хорошо ориентируется среди корабельных остовов и устраивает засады у выходов.','humanoid'),
 ('ord_mantell_pirate','Мантелльский пират','Скрывает награбленные грузы в прибрежных убежищах Орд-Мантелла. Охраняет тайные причалы и не оставляет свидетелей нелегальных разгрузок.','humanoid'),
 ('ord_mantell_scrapyard_droid','Сбойный утилизационный дроид','Разбирает старые машины с помощью тяжёлых промышленных инструментов. Повреждённые датчики больше не отличают живую цель от очередного объекта утилизации.','droid'),
 ('ord_mantell_savrip','Саврип-налётчик','Вооружённый саврип защищает скальные убежища и нападает на чужие стоянки. За внушительным обликом скрываются разум, наблюдательность и знание местных троп.','humanoid'),
 ('athiss_jungle_stalker','Хищник Атисса','Охотится среди корней и заросших каменных стен. Долго остаётся неподвижным в укрытии, пока добыча не приблизится к узкому проходу.','beast'),
 ('athiss_alchemical_beast','Алхимический зверь','Потомок существ, изменённых древними опытами ситхов. Вырвавшиеся из загонов звери заняли пустующие залы и подземные проходы храма.','beast'),
 ('athiss_temple_zealot','Фанатик древнего храма','Следует уцелевшим обрядам изолированного культа. Считает храмовые залы священными и встречает чужаков как осквернителей.','humanoid'),
 ('athiss_cult_overseer','Надзиратель храмового культа','Управляет стражей и присматривает за звериными загонами. Знает скрытые пути к внутренним помещениям и охраняет доступ к ним.','humanoid'),
 ('thule_steppe_predator','Степной хищник Туле','Следует за добычей по сухим руслам и высокой траве. Прячется от гроз в каменных укрытиях и защищает их от других обитателей степи.','beast'),
 ('thule_ruin_marauder','Мародёр руин Туле','Превращает древние укрепления в склады добычи. Разбирается в обходных путях и устраивает засады у спусков в подземелья.','humanoid'),
 ('thule_storm_cultist','Служитель грозового культа','Проводит обряды у древних башен Туле. Преследует тех, кто приближается к ритуальным площадкам без разрешения его общины.','humanoid'),
 ('thule_war_automaton','Боевой автомат Туле','Тяжёлая машина сохранила приказы защитников древнего бастиона. Пробуждается при нарушении периметра и не различает нынешние фракционные знаки.','droid'),
 ('ossus_dust_stalker','Пустошный хищник Оссуса','Приспособился к скудной жизни среди пыльных равнин и скальных трещин. Охотится вблизи укрытий, где путники пережидают бури.','beast'),
 ('ossus_relic_poacher','Охотник за реликвиями','Грабит древние хранилища ради покупателей на чёрном рынке. Защищает находки от любых соперников и использует руины для засад.','humanoid'),
 ('ossus_survey_droid','Повреждённый разведывательный дроид','Остался от одной из исчезнувших исследовательских групп. Продолжает обследовать руины, но сбои в памяти заставляют его атаковать приближающихся существ.','droid'),
 ('ossus_archive_sentry','Древний архивный страж','Охраняет уцелевшие секции хранилищ Оссуса. Системы допуска давно утратили связь с владельцами и не признают ни одного действующего разрешения.','droid')
ON CONFLICT (code) DO NOTHING;

INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,
       LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('arkania','snowbound_landing','arkania_tundra_stalker','underworld_scavenger','arkania_corporate_mercenary'),
 ('arkania','tundra_trails','arkania_tundra_stalker','arkania_tundra_stalker','arkania_corporate_mercenary'),
 ('arkania','ice_gorges','arkania_tundra_stalker','cave_predator','arkania_tundra_stalker'),
 ('arkania','mining_camp','underworld_scavenger','arkania_corporate_mercenary','damaged_drill_droid'),
 ('arkania','diamond_mines','rogue_loader','arkania_corporate_mercenary','damaged_drill_droid'),
 ('arkania','mining_tunnels','damaged_drill_droid','cave_predator','arkania_lab_droid'),
 ('arkania','research_outskirts','arkania_corporate_mercenary','arkania_lab_droid','arkania_corporate_mercenary'),
 ('arkania','sealed_laboratories','arkania_lab_droid','arkania_corporate_mercenary','arkania_lab_droid'),
 ('arkania','glacier_caverns','cave_predator','arkania_tundra_stalker','arkania_dragon'),
 ('arkania','dragon_cliffs','arkania_tundra_stalker','arkania_dragon','arkania_dragon'),
 ('balmorra','supply_platforms','underworld_scavenger','balmorra_factory_saboteur','balmorra_security_droid'),
 ('balmorra','workshop_streets','rogue_loader','balmorra_factory_saboteur','balmorra_security_droid'),
 ('balmorra','scrap_yards','balmorra_wasteland_scavenger','underworld_scavenger','rogue_loader'),
 ('balmorra','assembly_district','rogue_loader','balmorra_factory_saboteur','balmorra_security_droid'),
 ('balmorra','freight_tunnels','contraband_guard','balmorra_factory_saboteur','balmorra_security_droid'),
 ('balmorra','chemical_wastes','balmorra_wasteland_scavenger','balmorra_wasteland_scavenger','balmorra_security_droid'),
 ('balmorra','abandoned_foundry','rogue_loader','balmorra_security_droid','damaged_drill_droid'),
 ('balmorra','weapons_warehouses','contraband_guard','balmorra_factory_saboteur','balmorra_assault_prototype'),
 ('balmorra','testing_grounds','balmorra_security_droid','balmorra_factory_saboteur','balmorra_assault_prototype'),
 ('balmorra','prototype_vault','balmorra_security_droid','balmorra_assault_prototype','balmorra_assault_prototype'),
 ('ord_mantell','cargo_quay','ord_mantell_scrap_raider','contraband_guard','ord_mantell_pirate'),
 ('ord_mantell','coastal_tracks','ord_mantell_scrap_raider','ord_mantell_pirate','ord_mantell_savrip'),
 ('ord_mantell','junk_market','ord_mantell_scrap_raider','gang_enforcer','gang_leader'),
 ('ord_mantell','wreck_fields','ord_mantell_scrap_raider','ord_mantell_scrapyard_droid','ord_mantell_scrapyard_droid'),
 ('ord_mantell','scavenger_camps','ord_mantell_scrap_raider','ord_mantell_scrapyard_droid','gang_leader'),
 ('ord_mantell','lava_ravines','ord_mantell_savrip','ord_mantell_scrap_raider','ord_mantell_savrip'),
 ('ord_mantell','smugglers_coves','contraband_guard','ord_mantell_pirate','ord_mantell_pirate'),
 ('ord_mantell','pirate_docks','ord_mantell_pirate','ord_mantell_scrapyard_droid','gang_leader'),
 ('ord_mantell','savrip_highlands','ord_mantell_savrip','ord_mantell_savrip','ord_mantell_savrip'),
 ('ord_mantell','raider_stronghold','ord_mantell_scrap_raider','ord_mantell_pirate','gang_leader'),
 ('athiss','expedition_landing','athiss_jungle_stalker','tomb_raider','athiss_temple_zealot'),
 ('athiss','overgrown_trail','athiss_jungle_stalker','athiss_jungle_stalker','athiss_temple_zealot'),
 ('athiss','vine_gorge','athiss_jungle_stalker','tomb_raider','athiss_alchemical_beast'),
 ('athiss','sunken_ruins','tomb_raider','athiss_jungle_stalker','athiss_alchemical_beast'),
 ('athiss','broken_colonnade','athiss_temple_zealot','guardian_droid','athiss_cult_overseer'),
 ('athiss','temple_courtyard','athiss_temple_zealot','athiss_alchemical_beast','athiss_cult_overseer'),
 ('athiss','ritual_halls','athiss_temple_zealot','guardian_droid','athiss_cult_overseer'),
 ('athiss','alchemical_chambers','athiss_alchemical_beast','athiss_temple_zealot','athiss_cult_overseer'),
 ('athiss','beast_pits','athiss_alchemical_beast','athiss_cult_overseer','athiss_alchemical_beast'),
 ('athiss','buried_sanctum','athiss_temple_zealot','athiss_alchemical_beast','athiss_cult_overseer'),
 ('thule','dry_steppe','thule_steppe_predator','thule_ruin_marauder','thule_steppe_predator'),
 ('thule','storm_pass','thule_steppe_predator','thule_ruin_marauder','thule_storm_cultist'),
 ('thule','basalt_gorges','thule_steppe_predator','cave_predator','thule_ruin_marauder'),
 ('thule','abandoned_waystation','thule_ruin_marauder','rogue_sentry','thule_war_automaton'),
 ('thule','ruin_fields','tomb_raider','thule_ruin_marauder','thule_storm_cultist'),
 ('thule','cracked_aqueducts','thule_ruin_marauder','thule_steppe_predator','guardian_droid'),
 ('thule','underground_cisterns','thule_ruin_marauder','guardian_droid','thule_war_automaton'),
 ('thule','sith_bastion','thule_storm_cultist','thule_war_automaton','thule_war_automaton'),
 ('thule','lightning_spires','thule_storm_cultist','guardian_droid','thule_storm_cultist'),
 ('thule','sealed_war_vault','guardian_droid','thule_storm_cultist','thule_war_automaton'),
 ('ossus','dusty_landing','ossus_dust_stalker','ossus_relic_poacher','ossus_survey_droid'),
 ('ossus','ash_plains','ossus_dust_stalker','ossus_relic_poacher','ossus_dust_stalker'),
 ('ossus','wreckage_hollows','ossus_relic_poacher','ossus_survey_droid','rogue_loader'),
 ('ossus','ruined_avenues','ossus_relic_poacher','ossus_dust_stalker','ossus_survey_droid'),
 ('ossus','rock_shelters','ossus_dust_stalker','ossus_relic_poacher','cave_predator'),
 ('ossus','buried_archives','ossus_relic_poacher','ossus_survey_droid','ossus_archive_sentry'),
 ('ossus','library_approach','ossus_relic_poacher','ossus_archive_sentry','ossus_archive_sentry'),
 ('ossus','ruined_galleries','ossus_survey_droid','ossus_relic_poacher','ossus_archive_sentry'),
 ('ossus','sealed_stacks','ossus_archive_sentry','ossus_relic_poacher','ossus_archive_sentry'),
 ('ossus','deep_repository','ossus_archive_sentry','ossus_survey_droid','ossus_archive_sentry')
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

-- Permanent shared slots; neither rank edits nor rereads create extra instances.
INSERT INTO npc_spawns(location_npc_id,slot)
SELECT n.id,slots.slot FROM location_npcs n
JOIN locations l ON l.id=n.location_id
JOIN planets p ON p.id=l.planet_id
CROSS JOIN LATERAL generate_series(1,CASE n.code WHEN 'normal' THEN 2 ELSE 1 END) AS slots(slot)
WHERE p.code IN ('arkania','balmorra','ord_mantell','athiss','thule','ossus') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
