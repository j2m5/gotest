-- First additional protected-core batch. Core membership/control stay in seed07.
-- Geography inspires authored districts and enemy roles, not exact historical borders.
-- Early Great Galactic War: no later Alderaan succession war, occupied Corellia,
-- devastated Ziost, named Cold War estates on Begeren or later Bosthirda prophets.
-- Kuat locations are on the surface; orbital construction is supplied from these sites.

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('corellia_dock_raider','Налётчик грузовых кварталов','Промышляет кражами деталей и нападениями на перевозчиков Кореллии. Скрывает добычу в закрытых цехах и защищает нелегальные перевалочные пункты.','humanoid'),
 ('corellia_industrial_spy','Вооружённый промышленный шпион','Похищает чертежи и заводские допуски по заказу нелегальных посредников. Готов силой удерживать захваченные архивы и устранять свидетелей своих операций.','humanoid'),
 ('corellia_assembly_droid','Сбойный сборочный дроид','Остался в покинутом кореллианском цехе без надёжного управления. Повреждённая программа заставляет его обрабатывать любой движущийся объект как деталь конвейера.','droid'),
 ('corellia_feral_guard_beast','Одичавший сторожевой зверь','Сбежал из подпольного питомника и укрылся в городских зарослях. Сохранил привычку нападать на незнакомцев, но больше не подчиняется командам хозяев.','beast'),
 ('alderaan_mountain_predator','Альдераанский горный хищник','Держится у лесных границ и защищённых от ветра склонов. Выслеживает добычу у ручьёв и старых троп, ведущих через горные владения.','beast'),
 ('alderaan_poacher','Браконьер горных владений','Незаконно охотится в удалённых лесах Альдераана. Скрывает ловушки и добычу среди покинутых хозяйственных построек, защищая свой промысел оружием.','humanoid'),
 ('alderaan_rogue_retainer','Наёмник захваченной усадьбы','Присоединился к группе, присвоившей имущество пустующего владения. Контролирует дороги и охраняет склады награбленного в старых жилых корпусах.','humanoid'),
 ('alderaan_estate_droid','Взломанный усадебный дроид','Бывший охранник закрытой резиденции получил чужие коды управления. Теперь защищает укрывшихся внутри преступников и не признаёт законные пропуска.','droid'),
 ('kuat_cargo_hijacker','Похититель куатских грузов','Перехватывает поставки оборудования для подрядчиков верфей. Пользуется поддельными ведомостями и прячет украденные узлы в закрытых наземных складах.','humanoid'),
 ('kuat_industrial_saboteur','Промышленный диверсант','Проникает в испытательные помещения и архивы Куата. Готовит опасные подмены оборудования и охраняет тайники с украденной документацией.','humanoid'),
 ('kuat_service_droid','Неисправный сервисный дроид','Обслуживал наземные инженерные сети и грузовые комплексы. Утратив связь с диспетчером, применяет рабочие инструменты против движущихся препятствий.','droid'),
 ('kuat_security_droid','Перепрограммированный автомат охраны','Списанная охранная машина возвращена к работе нелегальными хозяевами. Блокирует проходы в присвоенных комплексах и атакует посетителей без их допуска.','droid'),
 ('ziost_frost_stalker','Хищник каменной тундры','Прячется в защищённых от холодного ветра впадинах Зиоста. Выходит к дорогам и пещерам, когда поблизости появляется подходящая добыча.','beast'),
 ('ziost_relic_raider','Зиостский расхититель реликвий','Проникает в закрытые хранилища и заброшенные древние сооружения. Охраняет самовольные раскопки и тайники, стараясь не привлекать законные власти.','humanoid'),
 ('ziost_sentinel_droid','Повреждённый страж Зиоста','Механический часовой продолжает службу у давно закрытого комплекса. Старые протоколы не распознают современные допуски и считают любое приближение вторжением.','droid'),
 ('ziost_renegade_adept','Отступник древних залов','Ищет силу в запретных реликвиях Зиоста вне надзора наставников. Подчинил себе часть старой охраны и преследует тех, кто обнаружил его убежище.','humanoid'),
 ('begeren_desert_stalker','Хищник скальных пустынь','Обитает среди каменных гряд Бегерена. Использует глубокие тени у обрывов и выработок, чтобы незаметно приблизиться к добыче.','beast'),
 ('begeren_gem_raider','Похититель самоцветов Бегерена','Присваивает добычу небольших экспедиций и грабит удалённые склады. Защищает незаконные выработки вместе с наёмниками и переделанной техникой.','humanoid'),
 ('begeren_mining_droid','Повреждённый самоцветный бурильщик','Добывающая машина осталась без обслуживания в старых карьерах. Продолжает дробить и перемещать породу, принимая приближающихся существ за рабочие цели.','droid'),
 ('begeren_monument_guardian','Механический страж монументов','Охранный автомат сохранился у древних скальных сооружений. Повреждённая система опознавания мешает ему отличить разрешённого посетителя от расхитителя.','droid'),
 ('bosthirda_cliff_stalker','Скальный хищник Бостхирды','Двигается между естественными нишами и пещерами горного хребта. Нападает на тех, кто проходит слишком близко к его укрытию или водопою.','beast'),
 ('bosthirda_pass_raider','Разбойник горных перевалов','Перехватывает обозы на удалённых дорогах Бостхирды. Использует старые наблюдательные посты для засад и охраны награбленных припасов.','humanoid'),
 ('bosthirda_renegade_acolyte','Послушник горного убежища','Скрывается от имперского надзора в заброшенных залах Бостхирды. Собирает запрещённые находки и держит вооружённых сообщников у входов в своё укрытие.','humanoid'),
 ('bosthirda_sentry_droid','Сбойный горный часовой','Старая охранная машина патрулирует каменные переходы и закрытые ворота. Изменённые или повреждённые команды заставляют её атаковать любого постороннего.','droid')
ON CONFLICT (code) DO NOTHING;

-- Republic core: Corellia, Alderaan and Kuat.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('corellia','trade_port_apron','Площадка торгового порта','Грузовые опоры стоят между пассажирскими терминалами и складскими рядами. За пределами людных проходов скрываются похитители небольших торговых партий.',1,1),
 ('corellia','freight_corridors','Грузовые коридоры','Высокие транспортные эстакады пересекают старые служебные кварталы. Нижние проезды заняты нелегальными мастерскими и стоянками перехваченных платформ.',2,6),
 ('corellia','overgrown_park','Заросший городской парк','Среди промышленных зданий сохранился большой зелёный участок. В закрытых для прогулок зарослях прячутся сбежавшие сторожевые звери и вооружённые браконьеры.',3,11),
 ('corellia','old_transit_tunnels','Старые транспортные тоннели','Неиспользуемые ветки соединяют подвалы складов с заводскими окраинами. По ним перемещают краденые детали, обходя основные пункты досмотра.',4,16),
 ('corellia','closed_assembly_district','Закрытый сборочный район','Пустующие цехи отделены от рабочих предприятий высоким ограждением. Внутри снова запущены некоторые линии, но их хозяева не значатся в заводских реестрах.',5,21),
 ('corellia','riverside_warehouses','Приречные склады','Старые причалы примыкают к грузовым ангарам на речном берегу. Через боковые ворота вывозят оборудование, украденное с действующих производств.',6,26),
 ('corellia','cargo_interchange','Грузовая пересадочная станция','Несколько ярусов путей сходятся у заброшенного перегрузочного узла. Налётчики превратили пустые диспетчерские в наблюдательные посты.',7,31),
 ('corellia','design_archive','Закрытый конструкторский архив','Сейфовые комнаты старого проектного бюро соединены служебными переходами. Промышленные шпионы захватили часть помещений и отключили внешние каналы связи.',8,36),
 ('corellia','abandoned_test_range','Покинутый испытательный полигон','За защитными валами расположены стенды и изолированные ангары. Оставленные машины продолжают испытания по повреждённым программам, реагируя на любое движение.',9,41),
 ('corellia','illegal_ship_workshop','Подпольный корабельный цех','Под перекрытиями большого ангара собирают суда из похищенных узлов. Вооружённые хозяева мастерской усилили охрану переделанными промышленными машинами.',10,46),
 ('alderaan','valley_landing','Посадочная площадка в долине','Каменные опоры расположены между лесистыми склонами и широкой дорогой. Дальние тропы уходят к охотничьим угодьям и небольшим горным владениям.',1,1),
 ('alderaan','forest_boundary','Лесная межа','Просека разделяет ухоженные владения и густой горный лес. На забытых ответвлениях встречаются браконьерские ловушки и следы диких зверей.',2,6),
 ('alderaan','snowmelt_streams','Ручьи талых снегов','Холодные потоки сбегают со склонов к низким каменным мостам. У водопоев держатся хищники, а в рощах скрыты стоянки нелегальных охотников.',3,11),
 ('alderaan','abandoned_hunting_lodge','Покинутый охотничий дом','Старые хозяйственные постройки стоят у края запущенного парка. Бывшие наёмные служащие присвоили имущество и силой удерживают подступы к усадьбе.',4,16),
 ('alderaan','highland_pass','Высокогорный перевал','Дорога поднимается между снежными карнизами и острыми скалами. Каменные укрытия заняты разбойниками, перехватывающими небольшие обозы.',5,21),
 ('alderaan','old_estate_gardens','Сады старого владения','Заросшие аллеи ведут к закрытым воротам пустующей резиденции. Между павильонами патрулируют дроиды, чьи коды доступа изменили незаконные хозяева.',6,26),
 ('alderaan','cliff_galleries','Скальные галереи','За высокими стенами ущелья сохранились крытые переходы и лестницы. Они связывают тайные склады вооружённой группы с удалёнными охотничьими тропами.',7,31),
 ('alderaan','snowbound_ruins','Заснеженные руины','Остатки каменных залов выступают над занесённой снегом террасой. Среди обвалов устроили логова звери, а уцелевшие подсобные камеры заняли расхитители.',8,36),
 ('alderaan','seized_gatehouse','Захваченный привратный корпус','Укреплённые ворота перекрывают дорогу к заброшенной резиденции. Банда бывших наёмников разместила здесь арсенал и перепрограммировала старую охрану.',9,41),
 ('alderaan','mountain_residence','Горная резиденция','На высокой скале стоит отрезанный от дорог усадебный комплекс. Присвоившие его преступники охраняют внутренние дворы и хранилища награбленного.',10,46),
 ('kuat','supply_landing','Посадочная площадка снабжения','Грузовой терминал принимает поставки для наземных служб и орбитальных верфей. За освещёнными воротами лежат старые складские секции с редкими патрулями.',1,1),
 ('kuat','service_boulevards','Служебные бульвары','Зелёные полосы отделяют дороги снабжения от аккуратных жилых кварталов. В закрытых технических дворах скрываются похитители грузов и поддельные перевозчики.',2,6),
 ('kuat','garden_maintenance','Садовые технические участки','Под ухоженными террасами расположены насосы и сервисные туннели. В отключённых от общей сети секциях продолжают работать машины с нарушенными программами.',3,11),
 ('kuat','component_depots','Склады корабельных узлов','Длинные ангары содержат разобранные механизмы и запасные блоки. Часть складов захвачена группой, подменяющей учётные записи и вывозящей ценные детали.',4,16),
 ('kuat','closed_transport_hub','Закрытый транспортный узел','Законсервированные погрузочные пути связывают несколько наземных терминалов. Нелегальные перевозчики восстановили местные подъёмники и установили собственные пропуска.',5,21),
 ('kuat','abandoned_test_labs','Покинутые испытательные лаборатории','За прочными перегородками сохранились стенды проверки корабельного оборудования. Диверсанты используют их для подготовки повреждённых узлов и опасных подмен.',6,26),
 ('kuat','estate_service_wing','Служебное крыло старой усадьбы','Заросший боковой двор скрывает переходы к пустующим хозяйственным корпусам. Похитители промышленных секретов устроили внутри охраняемые тайники.',7,31),
 ('kuat','ground_control_archive','Архив наземной диспетчерской','Сейфовые помещения хранят старые схемы снабжения и служебные допуски. Группа взломщиков удерживает архив, защищая доступ к сведениям о ценных поставках.',8,36),
 ('kuat','security_service_depot','Склад охранной техники','В изолированных ангарах собраны списанные автоматы защиты. Часть машин вернулась к работе под управлением людей, незаконно занявших комплекс.',9,41),
 ('kuat','sealed_supply_compound','Закрытый комплекс снабжения','Высокие ограждения отделяют складские дворы от остальной транспортной сети. Внутри укрыта техника и документация, похищенная у подрядчиков орбитальных верфей.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Imperial core: Ziost, Begeren and Bosthirda.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('ziost','cold_landing','Холодная посадочная площадка','За низкими стенами терминала тянутся промёрзшие каменные равнины. Старые дороги ведут к окраинным складам и давно закрытым комплексам.',1,1),
 ('ziost','stone_tundra','Каменистая тундра','Редкая растительность цепляется за защищённые от ветра впадины. У скальных выступов сходятся следы зверей и пути нелегальных искателей древностей.',2,6),
 ('ziost','ice_pass','Ледяной проход','Узкая дорога огибает покрытые льдом уступы. В расщелинах прячутся хищники, а под навесами стоят оставленные экспедиционные платформы.',3,11),
 ('ziost','abandoned_outskirts','Покинутые окраины','Ряды пустых хозяйственных зданий примыкают к старой дороге. Банды расхитителей разбирают уцелевшее оборудование и охраняют подступы к своим складам.',4,16),
 ('ziost','frozen_caverns','Промёрзшие пещеры','Лёд покрыл стены природных полостей под каменным плато. В боковых ходах скрываются звери, а в глубине видны следы незарегистрированных раскопок.',5,21),
 ('ziost','forgotten_archive','Забытый архив','Каменные залы старого хранилища защищены тяжёлыми дверями. Искатели запретных знаний проникли внутрь и восстановили часть древних охранных цепей.',6,26),
 ('ziost','wind_towers','Башни на ветру','Разрушенные опоры возвышаются над холодной равниной. На уцелевших площадках расположены наблюдательные посты и скрытые запасы расхитителей.',7,31),
 ('ziost','sealed_crypts','Закрытые погребальные залы','Под массивными плитами сохранились узкие коридоры древнего некрополя. Самовольные исследователи охраняют найденные реликвии и не допускают сюда свидетелей.',8,36),
 ('ziost','ancient_causeway','Древняя каменная дамба','Широкий переход связывает старые укрепления над обледеневшей низиной. Повреждённые механические стражи продолжают службу у перекрытых ворот.',9,41),
 ('ziost','old_citadel_court','Двор старой цитадели','Внутренние площадки давно выведенного из службы укрепления закрыты от ветра высокими стенами. Отступники превратили уцелевшие залы в убежище и подчинили себе старую охрану.',10,46),
 ('begeren','desert_landing','Пустынная посадочная площадка','Грузовые опоры стоят между каменными грядами и дорогой к карьерам. В стороне от рабочих маршрутов лежат покинутые склады добывающих экспедиций.',1,1),
 ('begeren','gemstone_flats','Самоцветные пустоши','Светлые скопления минералов выступают из сухой породы. В складках рельефа прячутся хищники и искатели, готовые силой отстаивать найденные участки.',2,6),
 ('begeren','cliff_quarry','Карьер у скальных обрывов','Широкие уступы уходят под нависающую каменную стену. Списанные буровые машины мешают нелегальным добытчикам, а те не терпят новых соперников.',3,11),
 ('begeren','broken_statues','Разбитые монументы','Обломки огромных изваяний лежат среди осыпей и старых выработок. Расхитители ищут сохранившиеся детали, используя каменные головы и постаменты как укрытия.',4,16),
 ('begeren','dry_extraction_tunnels','Сухие добывающие тоннели','Штольни уходят под древние скальные сооружения. В них ещё движется неисправная техника, а у боковых выходов устроены тайники похитителей самоцветов.',5,21),
 ('begeren','abandoned_sorting_yard','Покинутый сортировочный двор','Ряды дробилок и контейнеров занимают ровную площадку у карьера. Захватившая двор группа скупает незаконную добычу и охраняет погрузочные линии.',6,26),
 ('begeren','monument_terraces','Террасы монументов','Каменные ступени поднимаются к уцелевшим фигурам древних ситхов. Повреждённые охранные машины следят за подходами, не признавая современные допуски.',7,31),
 ('begeren','buried_mineral_vault','Засыпанное минеральное хранилище','Старые складские камеры скрыты под толстым слоем обломков. Нелегальные раскопщики восстановили несколько проходов и установили заслоны у найденных запасов.',8,36),
 ('begeren','cliff_watchworks','Скальные сторожевые сооружения','Узкие переходы соединяют охранные площадки высоко над выработками. Отсюда вооружённые хозяева контролируют дороги к тайным складам.',9,41),
 ('begeren','stone_sanctuary','Каменное святилище','В глубине древнего комплекса сохранились широкие залы и остатки монументальных изваяний. Расхитители удерживают входы, а внутри продолжают действовать старые механические стражи.',10,46),
 ('bosthirda','mountain_landing','Горная посадочная площадка','Небольшой терминал расположен у подножия высоких тёмных склонов. Дороги расходятся к поселениям, перевалам и старым укреплённым участкам.',1,1),
 ('bosthirda','foothill_tracks','Предгорные тропы','Каменистые дороги петляют между редкими укрытиями и глубокими ложбинами. На удалённых поворотах скрываются разбойники, следящие за небольшими обозами.',2,6),
 ('bosthirda','cliff_hollows','Скальные укрытия','Цепочка естественных ниш тянется вдоль горного уступа. Между логовами хищников встречаются забытые экспедиционные стоянки и тайники налётчиков.',3,11),
 ('bosthirda','old_watchposts','Старые наблюдательные посты','Низкие каменные башни контролировали несколько горных дорог. Теперь их занимают вооружённые группы, присвоившие уцелевшие запасы и сигнальное оборудование.',4,16),
 ('bosthirda','hidden_supply_route','Скрытая дорога снабжения','Узкий обход проходит за крутыми склонами вдали от основных трактов. Замаскированные склады позволяют налётчикам поддерживать свои удалённые убежища.',5,21),
 ('bosthirda','abandoned_retreat','Заброшенное горное убежище','Пустые жилые залы вырублены в защищённой от ветра скале. Здесь обосновались отступники, скрывающие свои находки и занятия от имперского надзора.',6,26),
 ('bosthirda','ridge_caverns','Пещеры горного хребта','Сеть широких полостей соединяет несколько ущелий. Её дальние камеры заняты хищниками и подпольными хранилищами вооружённых групп.',7,31),
 ('bosthirda','sealed_shrine','Закрытое святилище','Тяжёлые плиты отделяют внутренние залы от занесённых пылью площадок. Самовольные искатели силы удерживают входы и используют старые машины для охраны своих тайн.',8,36),
 ('bosthirda','mountain_redoubt','Укрепление на перевале','Баррикады перекрывают подходы к высокому каменному двору. Отсюда отряд налётчиков контролирует обходные дороги и снабжает скрытые базы в горах.',9,41),
 ('bosthirda','hidden_highland_keep','Скрытая высокогорная крепость','За цепью узких проходов расположен давно покинутый укреплённый комплекс. Отступники и их наёмники восстановили внутреннюю охрану, закрыв доступ к своим складам и залам.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('corellia','trade_port_apron','corellia_dock_raider','rogue_loader','corellia_assembly_droid'),
 ('corellia','freight_corridors','corellia_dock_raider','corellia_assembly_droid','corellia_industrial_spy'),
 ('corellia','overgrown_park','corellia_feral_guard_beast','corellia_dock_raider','corellia_feral_guard_beast'),
 ('corellia','old_transit_tunnels','corellia_dock_raider','corellia_feral_guard_beast','corellia_industrial_spy'),
 ('corellia','closed_assembly_district','corellia_assembly_droid','corellia_dock_raider','corellia_assembly_droid'),
 ('corellia','riverside_warehouses','corellia_dock_raider','rogue_loader','corellia_industrial_spy'),
 ('corellia','cargo_interchange','corellia_dock_raider','corellia_assembly_droid','corellia_industrial_spy'),
 ('corellia','design_archive','corellia_industrial_spy','rogue_sentry','corellia_industrial_spy'),
 ('corellia','abandoned_test_range','corellia_assembly_droid','rogue_sentry','corellia_assembly_droid'),
 ('corellia','illegal_ship_workshop','corellia_dock_raider','corellia_industrial_spy','corellia_assembly_droid'),
 ('alderaan','valley_landing','alderaan_poacher','alderaan_mountain_predator','alderaan_rogue_retainer'),
 ('alderaan','forest_boundary','alderaan_mountain_predator','alderaan_poacher','alderaan_mountain_predator'),
 ('alderaan','snowmelt_streams','alderaan_mountain_predator','alderaan_poacher','alderaan_poacher'),
 ('alderaan','abandoned_hunting_lodge','alderaan_poacher','alderaan_rogue_retainer','alderaan_estate_droid'),
 ('alderaan','highland_pass','alderaan_mountain_predator','alderaan_poacher','alderaan_rogue_retainer'),
 ('alderaan','old_estate_gardens','alderaan_poacher','alderaan_estate_droid','alderaan_rogue_retainer'),
 ('alderaan','cliff_galleries','alderaan_rogue_retainer','alderaan_estate_droid','alderaan_rogue_retainer'),
 ('alderaan','snowbound_ruins','alderaan_mountain_predator','alderaan_poacher','alderaan_estate_droid'),
 ('alderaan','seized_gatehouse','alderaan_rogue_retainer','alderaan_estate_droid','alderaan_estate_droid'),
 ('alderaan','mountain_residence','alderaan_rogue_retainer','alderaan_estate_droid','alderaan_rogue_retainer'),
 ('kuat','supply_landing','kuat_cargo_hijacker','kuat_service_droid','kuat_industrial_saboteur'),
 ('kuat','service_boulevards','kuat_cargo_hijacker','kuat_industrial_saboteur','kuat_service_droid'),
 ('kuat','garden_maintenance','kuat_service_droid','kuat_industrial_saboteur','kuat_service_droid'),
 ('kuat','component_depots','kuat_cargo_hijacker','kuat_service_droid','kuat_security_droid'),
 ('kuat','closed_transport_hub','kuat_cargo_hijacker','kuat_service_droid','kuat_security_droid'),
 ('kuat','abandoned_test_labs','kuat_industrial_saboteur','kuat_service_droid','kuat_security_droid'),
 ('kuat','estate_service_wing','kuat_cargo_hijacker','kuat_industrial_saboteur','kuat_security_droid'),
 ('kuat','ground_control_archive','kuat_industrial_saboteur','kuat_security_droid','kuat_industrial_saboteur'),
 ('kuat','security_service_depot','kuat_security_droid','kuat_service_droid','kuat_security_droid'),
 ('kuat','sealed_supply_compound','kuat_cargo_hijacker','kuat_industrial_saboteur','kuat_security_droid'),
 ('ziost','cold_landing','ziost_relic_raider','ziost_frost_stalker','ziost_sentinel_droid'),
 ('ziost','stone_tundra','ziost_frost_stalker','ziost_relic_raider','ziost_frost_stalker'),
 ('ziost','ice_pass','ziost_frost_stalker','ziost_relic_raider','ziost_sentinel_droid'),
 ('ziost','abandoned_outskirts','ziost_relic_raider','ziost_sentinel_droid','ziost_relic_raider'),
 ('ziost','frozen_caverns','cave_predator','ziost_frost_stalker','ziost_relic_raider'),
 ('ziost','forgotten_archive','ziost_relic_raider','ziost_sentinel_droid','ziost_renegade_adept'),
 ('ziost','wind_towers','ziost_relic_raider','ziost_sentinel_droid','ziost_renegade_adept'),
 ('ziost','sealed_crypts','ziost_relic_raider','ziost_renegade_adept','ziost_sentinel_droid'),
 ('ziost','ancient_causeway','ziost_sentinel_droid','ziost_relic_raider','ziost_sentinel_droid'),
 ('ziost','old_citadel_court','ziost_renegade_adept','ziost_sentinel_droid','ziost_renegade_adept'),
 ('begeren','desert_landing','begeren_gem_raider','begeren_desert_stalker','begeren_mining_droid'),
 ('begeren','gemstone_flats','begeren_desert_stalker','begeren_gem_raider','begeren_desert_stalker'),
 ('begeren','cliff_quarry','begeren_gem_raider','begeren_mining_droid','begeren_mining_droid'),
 ('begeren','broken_statues','begeren_desert_stalker','begeren_gem_raider','begeren_monument_guardian'),
 ('begeren','dry_extraction_tunnels','begeren_gem_raider','begeren_desert_stalker','begeren_mining_droid'),
 ('begeren','abandoned_sorting_yard','begeren_gem_raider','begeren_mining_droid','begeren_gem_raider'),
 ('begeren','monument_terraces','begeren_monument_guardian','begeren_gem_raider','begeren_monument_guardian'),
 ('begeren','buried_mineral_vault','begeren_gem_raider','begeren_mining_droid','begeren_monument_guardian'),
 ('begeren','cliff_watchworks','begeren_gem_raider','begeren_monument_guardian','begeren_gem_raider'),
 ('begeren','stone_sanctuary','begeren_gem_raider','begeren_monument_guardian','begeren_monument_guardian'),
 ('bosthirda','mountain_landing','bosthirda_pass_raider','bosthirda_cliff_stalker','bosthirda_sentry_droid'),
 ('bosthirda','foothill_tracks','bosthirda_pass_raider','bosthirda_cliff_stalker','bosthirda_pass_raider'),
 ('bosthirda','cliff_hollows','bosthirda_cliff_stalker','bosthirda_pass_raider','bosthirda_cliff_stalker'),
 ('bosthirda','old_watchposts','bosthirda_pass_raider','bosthirda_sentry_droid','bosthirda_sentry_droid'),
 ('bosthirda','hidden_supply_route','bosthirda_pass_raider','bosthirda_sentry_droid','bosthirda_renegade_acolyte'),
 ('bosthirda','abandoned_retreat','bosthirda_renegade_acolyte','bosthirda_pass_raider','bosthirda_sentry_droid'),
 ('bosthirda','ridge_caverns','bosthirda_cliff_stalker','bosthirda_pass_raider','bosthirda_renegade_acolyte'),
 ('bosthirda','sealed_shrine','bosthirda_renegade_acolyte','bosthirda_sentry_droid','bosthirda_renegade_acolyte'),
 ('bosthirda','mountain_redoubt','bosthirda_pass_raider','bosthirda_sentry_droid','bosthirda_renegade_acolyte'),
 ('bosthirda','hidden_highland_keep','bosthirda_renegade_acolyte','bosthirda_pass_raider','bosthirda_sentry_droid')
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
WHERE p.code IN ('corellia','alderaan','kuat','ziost','begeren','bosthirda') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
