-- Second frontier batch. Initial/current ownership belongs to seed07/runtime.
-- Geography and local fauna draw on Legends; districts and enemy roles are authored.
-- Early Great Galactic War: no Palpatine-era Carida base, Trade Federation/B1s,
-- destroyed Livien Magnus, later Mirial buried-city discovery or replayed KOTORII.

-- Republic holdings: Carida, Cato Neimoidia and Shili.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('carida','cliffside_landing','Пригорная посадочная площадка','Грузовые суда садятся у основания высокого хребта. От площадки расходятся дороги к небольшим поселениям, старым складам и закрытым учебным районам.',1,1),
 ('carida','foothill_roads','Дороги предгорий','Каменные гряды разделяют извилистые пути между долинами. За поворотами встречаются брошенные машины, укрытия зверей и следы налётов на грузовые колонны.',2,6),
 ('carida','supply_depot','Старый снабженческий узел','Ряды ангаров окружают давно опустевшую распределительную площадку. В закрытых секциях остались запасы снаряжения, ради которых сюда возвращаются вооружённые группы.',3,11),
 ('carida','dry_canyons','Сухие каньоны','Ветер обтачивает узкие проходы между высокими стенами. Пересохшие русла ведут к пещерам и обходным тропам, скрытым от наблюдения с плато.',4,16),
 ('carida','abandoned_barracks','Покинутые казармы','Старые жилые корпуса стоят по краям заросшего плаца. Внутренние переходы соединяют оружейные комнаты, склады формы и помещения оставленных охранных машин.',5,21),
 ('carida','obstacle_range','Заброшенная полоса испытаний','Искусственные стены и рвы повторяют сложный горный рельеф. Часть учебных установок по-прежнему реагирует на движение, хотя испытателей здесь давно нет.',6,26),
 ('carida','mountain_rail','Горная грузовая ветка','Рельсы пересекают ущелья по укреплённым эстакадам. В тупиках остались составы с оборудованием, а под мостами устроены стоянки похитителей грузов.',7,31),
 ('carida','sealed_armory','Закрытые оружейные склады','За каменными валами скрыты складские бункеры. Повреждённые шлюзы открывают путь к оружию, которое охраняют наёмники и автономные системы.',8,36),
 ('carida','siege_range','Полигон тяжёлых машин','Испытательные позиции вырублены прямо в склоне. Между укреплёнными мишенями движутся тяжёлые автоматы, продолжающие отрабатывать давно забытые боевые задания.',9,41),
 ('carida','plateau_fortress','Крепость на плато','Массивные стены возвышаются над грузовыми дорогами и полигонами. Нижние ворота вскрыты, но внутренние дворы и машинные залы всё ещё удерживает опасная охрана.',10,46),
 ('cato_neimoidia','arrival_terraces','Приёмные террасы','Посадочные площадки пристроены к скальной опоре мостового города. Между грузовыми лифтами и залами досмотра тесно стоят контейнеры торговых домов.',1,1),
 ('cato_neimoidia','cargo_bridges','Грузовые мосты','Широкие мостовые пролёты соединяют перегрузочные терминалы. Под ними скрываются узкие служебные настилы, удобные для тайных перевозок и засад.',2,6),
 ('cato_neimoidia','lower_arcades','Нижние галереи','Торговые помещения нижнего яруса освещены отражённым светом тумана. Между закрытыми лавками проложены проходы к складам и неприметным частным лифтам.',3,11),
 ('cato_neimoidia','misty_cliffs','Туманные скалы','Неровные уступы спускаются от городских опор в густую пелену. На скальных полках сохранились заросли и звериные укрытия, почти незаметные с мостов.',4,16),
 ('cato_neimoidia','underbridge_walkways','Подмостовые переходы','Служебные дорожки оплетают нижнюю сторону городских конструкций. Сюда редко спускается обычная охрана, и шум транспорта скрывает шаги нарушителей.',5,21),
 ('cato_neimoidia','merchant_warehouses','Купеческие склады','За неприметными фасадами размещены многоярусные хранилища. Несколько секций захвачены нелегальными торговцами, которые сменили пропуска и перепрограммировали охрану.',6,26),
 ('cato_neimoidia','silent_pavilions','Безлюдные павильоны','Пустующие залы когда-то принимали богатых гостей. Под декоративными перекрытиями устроены тайники, а запертые боковые галереи ведут к частным сейфовым секциям.',7,31),
 ('cato_neimoidia','treasury_approach','Подступы к казначейству','Подвесные переходы сходятся у укреплённого финансового комплекса. Внешние посты покинуты, но новые баррикады показывают, что ценности внутри нашли других хозяев.',8,36),
 ('cato_neimoidia','suspended_vaults','Подвесные хранилища','Тяжёлые сейфовые модули закреплены между несущими опорами. Добраться до них можно по узким мосткам, простреливаемым из охранных ниш.',9,41),
 ('cato_neimoidia','sealed_treasury','Запечатанная сокровищница','В глубине комплекса сохранился изолированный зал с несколькими поясами защиты. Над вскрытыми внешними дверями горят сигналы тревоги, а внутренние механизмы продолжают охранять проходы.',10,46),
 ('shili','grassland_landing','Посадочная поляна','Расчищенная площадка лежит среди высоких разноцветных трав. Дальше от сигнальных огней начинаются звериные тропы и редкие пути местных охотников.',1,1),
 ('shili','colored_plains','Пёстрые равнины','Полосы трав разного оттенка тянутся до далёких холмов. Их плотные заросли скрывают движение небольших зверей и крупных хищников.',2,6),
 ('shili','forest_margin','Лесная окраина','Равнина сменяется редколесьем с густым подлеском. Под деревьями лежат следы браконьерских стоянок, а на ветвях закреплены забытые наблюдательные устройства.',3,11),
 ('shili','river_reeds','Речные заросли','Узкая река извивается между высокими травами и низкими деревьями. Удобные броды привлекают зверей, но рядом с водопоями расставлены чужие ловушки.',4,16),
 ('shili','poacher_camps','Лагеря браконьеров','Скрытые в складках местности площадки заставлены клетками и разобранным снаряжением. Между лагерями ходят вооружённые группы, охраняющие добычу для тайных покупателей.',5,21),
 ('shili','redgrass_ravines','Овраги красных трав','Глубокие ложбины почти исчезают под колышущейся красной травой. Со дна трудно увидеть приближение хищника, а выбраться на склон можно лишь в нескольких местах.',6,26),
 ('shili','ancient_groves','Старые рощи','Разросшиеся деревья образуют укрытые от ветра островки среди равнин. Между корнями сохранились старые тропы, которыми теперь пользуются звери и чужие охотничьи отряды.',7,31),
 ('shili','hidden_pens','Скрытые загоны','За растительными заслонами построены прочные клетки и погрузочные платформы. Ловчие машины сторожат периметр, а повреждённые ограждения выдают недавние попытки побега.',8,36),
 ('shili','akul_hunting_grounds','Охотничьи земли акулов','В высокой траве видны широкие проходы, проложенные крупными зверями. Останки добычи и глубокие следы когтей отмечают границы их охотничьих угодий.',9,41),
 ('shili','predator_plateau','Плато хищников','Изолированное плато окружено крутыми оврагами. Среди заросших каменных уступов устроены логова акулов, защищённые от большинства случайных посетителей.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

-- Imperial holdings: Ord Radama, Mirial and Telos IV. Control is a game setting;
-- local peoples and cultures are not presented as inherently Sith or hostile.
INSERT INTO locations(planet_id,code,name,description,position,recommended_level)
SELECT p.id,v.code,v.name,v.description,v.position,v.recommended_level FROM (VALUES
 ('ord_radama','marshside_quay','Пристань у болот','Деревянные и металлические настилы соединяют грузовые площадки с городскими дорогами. За последними опорами начинаются мелкие протоки и заросшие болотные островки.',1,1),
 ('ord_radama','canal_lanes','Улочки у каналов','Узкие улицы идут вдоль тёмной воды между тесно стоящими зданиями. Старые мостики и проходные дворы позволяют незаметно пересечь несколько кварталов.',2,6),
 ('ord_radama','damp_warehouses','Сырые склады','Нижние этажи припортовых зданий пропитаны влагой. В закрытых помещениях хранятся тайные грузы, а у боковых дверей дежурят вооружённые наблюдатели.',3,11),
 ('ord_radama','reed_fens','Тростниковые топи','Плотные заросли скрывают небольшие плёсы и полосы вязкой земли. Между редкими сухими кочками проходят звериные тропы и плохо заметные настилы.',4,16),
 ('ord_radama','old_waterworks','Старые водосбросы','Каменные затворы отделяют городские каналы от заболоченных низин. Часть механизмов работает сама по себе, а обслуживающие галереи заняты незваными обитателями.',5,21),
 ('ord_radama','flood_tunnels','Затопленные тоннели','Под кварталами тянутся полузатопленные дренажные коридоры. На сухих площадках спрятаны ящики, запасные батареи и следы недавних стоянок.',6,26),
 ('ord_radama','abandoned_residences','Заброшенные усадьбы','За заросшими оградами стоят опустевшие дома с внутренними садами. Их подвалы соединены со старыми водными ходами, которыми пользуются местные банды.',7,31),
 ('ord_radama','smugglers_islands','Островки контрабандистов','Мелкие острова среди протоков превращены в тайные перевалочные пункты. Под навесами спрятаны лодки и грузовые платформы, а подходы контролируют дозорные.',8,36),
 ('ord_radama','guild_depots','Склады тайных гильдий','Укреплённые склады занимают целый участок старого грузового района. Нелегальные объединения делят здесь товар, прикрывая сделки наёмной охраной и переделанными машинами.',9,41),
 ('ord_radama','canal_citadel','Цитадель у каналов','Старая водная крепость перекрывает несколько важных протоков. Вооружённая банда использует её дворы как защищённые причалы, а башни — для наблюдения за подступами.',10,46),
 ('mirial','cold_waystation','Холодный перевалочный пункт','Небольшой грузовой узел стоит на краю сухой равнины. Тяжёлые двери защищают помещения от ветра, а караванные пути расходятся к далёким горным укрытиям.',1,1),
 ('mirial','wind_scoured_plains','Продуваемые равнины','Ветер гонит сухую пыль над холодной каменистой землёй. Редкие расселины дают защиту от непогоды, привлекая путников и хищников.',2,6),
 ('mirial','caravan_gorge','Караванное ущелье','Путь между торговыми стоянками зажат отвесными стенами. В боковых проходах встречаются следы костров и спрятанное снаряжение тех, кто поджидает чужие грузы.',3,11),
 ('mirial','frost_caves','Пещеры инея','Холодный воздух оседает в глубоких каменных полостях. Тонкий иней покрывает стены, а в защищённых от ветра нишах устроены звериные логова.',4,16),
 ('mirial','abandoned_cisterns','Заброшенные водосборники','Старые резервуары собирали редкую влагу для караванных стоянок. Теперь в их служебных камерах скрываются налётчики и оставленные без обслуживания машины.',5,21),
 ('mirial','mountain_paths','Горные тропы','Каменные ступени поднимаются от пустошей к высоким уступам. Укрытия вдоль пути давно опустели, и лишь отдельные следы показывают, что тропами ещё пользуются.',6,26),
 ('mirial','deserted_retreat','Опустевший скит','Небольшие помещения высечены в защищённой от ветра скале. Их прежние обитатели ушли, оставив внутренние дворы группе отступников, закрывшей путь для чужаков.',7,31),
 ('mirial','sealed_cloisters','Закрытые переходы','Узкие галереи соединяют отдельные части горного убежища. На дверях сохранились старые знаки, а новые засовы и посты охраны выдают недавнее заселение.',8,36),
 ('mirial','cliff_sanctuaries','Скальные святилища','Уединённые каменные площадки расположены высоко над равниной. В прилегающих залах скрыты тайники отступников, охраняемые людьми и древними механизмами.',9,41),
 ('mirial','hidden_repository','Скрытое горное хранилище','За несколькими закрытыми проходами лежат глубокие складские камеры. Их содержимое привлекло новых хозяев, но часть древней охраны до сих пор никому не подчиняется.',10,46),
 ('telos_iv','lakeside_landing','Приозёрная посадочная площадка','Посадочные огни отражаются в воде большого озера. От грузовых навесов уходят дороги к лесным участкам и оставшимся от прежних работ техническим сооружениям.',1,1),
 ('telos_iv','restored_woodlands','Восстановленные леса','Деревья сомкнулись над старыми маршрутами обслуживания. В густом подлеске движутся потомки завезённых животных, а на стволах ещё видны выцветшие метки наблюдателей.',2,6),
 ('telos_iv','grassy_ravines','Травянистые овраги','Зелёные склоны спускаются к каменистым ручьям. Узкие проходы между зарослями служат звериными маршрутами и укрывают следы нелегальных стоянок.',3,11),
 ('telos_iv','survey_station','Старая наблюдательная станция','Изолированный комплекс следил за состоянием восстановленных земель. Большинство помещений опустело, но датчики и обслуживающие дроиды продолжают работать без единого диспетчера.',4,16),
 ('telos_iv','reservoir_shore','Берега водохранилища','Пологий берег чередуется с дамбами и каменными насыпями. Между заросшими техническими площадками прячутся животные и группы, разбирающие старое оборудование.',5,21),
 ('telos_iv','overgrown_access','Дороги под зарослями','Растения проросли через покрытие широких служебных дорог. На обочинах лежат разбитые транспортные платформы, а боковые съезды ведут к закрытым объектам.',6,26),
 ('telos_iv','deserted_service_hub','Покинутый технический узел','В бывшем центре обслуживания сохранились ремонтные боксы и линии энергоснабжения. Его оборудование привлекает налётчиков, которым мешают автономные машины.',7,31),
 ('telos_iv','buried_base','Погребённая база','Старые укрепления скрылись под грунтом и разросшейся зеленью. Через провалы в перекрытиях можно попасть в машинные залы, где ещё работает резервное питание.',8,36),
 ('telos_iv','control_bunkers','Управляющие бункеры','Низкие бронированные сооружения соединены подземными кабельными трассами. Сохранившиеся узлы передают противоречивые команды охране и ремонтным машинам.',9,41),
 ('telos_iv','sealed_defense_complex','Законсервированный оборонный комплекс','За тяжёлыми шлюзами остались ряды технических ниш и боевых машин. Нарушенные цепи консервации пробудили часть оборудования, превратив внутренние проходы в опасную зону.',10,46)
) AS v(planet_code,code,name,description,position,recommended_level) JOIN planets p ON p.code=v.planet_code
ON CONFLICT (planet_id,code) DO NOTHING;

INSERT INTO npc_templates(code,name,description,kind) VALUES
 ('carida_highland_predator','Горный хищник Кариды','Выслеживает добычу у каменных проходов и старых дорог. Скрывается в складках рельефа, выбирая момент, когда путнику трудно отступить.','beast'),
 ('carida_arms_raider','Похититель вооружения','Грабит склады снаряжения и перехватывает грузы в предгорьях Кариды. Использует найденные пропуска и военное оборудование для подготовки новых налётов.','humanoid'),
 ('carida_training_droid','Сбойный учебный дроид','Остался на закрытом участке старого полигона. Повреждённые ограничители превратили безопасные учебные задания в настоящую охоту на движущиеся цели.','droid'),
 ('carida_siege_automaton','Тяжёлый полигонный автомат','Массивная машина предназначалась для испытания укреплений. Утратила связь с управляющей станцией и продолжает выполнять последнюю программу штурма.','droid'),
 ('cato_neimoidia_bridge_raider','Налётчик мостовых городов','Знает служебные настилы и скрытые подъёмники Като-Неймодии. Нападает на небольшие грузы и исчезает под мостами до прибытия законной охраны.','humanoid'),
 ('cato_neimoidia_vault_guard','Наёмник закрытых хранилищ','Охраняет присвоенные ценности по заказу нелегальных торговцев. Контролирует узкие проходы к сейфовым секциям и не допускает туда свидетелей.','humanoid'),
 ('cato_neimoidia_security_droid','Хранитель сейфового зала','Тяжёлый дроид защиты сокровищницы с повреждённой системой допуска. Продолжает охранять двери, игнорируя действующие торговые пропуска.','droid'),
 ('cato_neimoidia_cliff_predator','Хищник туманных скал','Обитает на заросших уступах под мостовыми городами. Подбирается к добыче сквозь туман, используя трещины и узкие каменные полки.','beast'),
 ('shili_akul','Акул','Крупный четвероногий хищник с оранжевой шерстью бродит среди трав Шили. Его мощные челюсти и цепкое внимание делают приближение к логову опасным.','beast'),
 ('shili_grass_stalker','Равнинный хищник','Прячется в высокой пёстрой траве Шили. Пользуется звериными тропами и поджидает добычу у водопоев, стараясь не попадаться более крупным охотникам.','beast'),
 ('shili_poacher','Браконьер Шили','Ловит местных зверей для нелегальной продажи. Защищает скрытые загоны и уничтожает тех, кто обнаружил его стоянки или оборудование.','humanoid'),
 ('shili_trap_droid','Ловчий дроид','Переделанная машина браконьеров преследует и загоняет живые цели. Её программа не различает зверей и людей, оказавшихся внутри охраняемого участка.','droid'),
 ('ord_radama_marsh_stalker','Болотный хищник Орд-Радамы','Затаивается у протоков и заросших настилов. Находит путь между городскими окраинами и топями, следуя за животными и неосторожными путниками.','beast'),
 ('ord_radama_dock_smuggler','Контрабандист припортовых кварталов','Перевозит запрещённые грузы по каналам Орд-Радамы. Охраняет скрытые склады и нападает на тех, кто подошёл слишком близко к месту разгрузки.','humanoid'),
 ('ord_radama_canal_droid','Повреждённый канальный дроид','Обслуживал затворы и расчищал водные проходы. Изношенные датчики заставляют его принимать приближающихся существ за препятствия, требующие немедленного устранения.','droid'),
 ('ord_radama_gang_captain','Главарь портовой банды','Удерживает несколько причалов и нелегальных складов. Собирает вооружённых сообщников для защиты водных маршрутов и захвата чужих грузов.','humanoid'),
 ('mirial_frost_stalker','Хищник холодных пустошей','Находит добычу среди каменных укрытий Мириала. Сохраняет силы в защищённых от ветра нишах и нападает, когда путник останавливается на отдых.','beast'),
 ('mirial_caravan_raider','Налётчик караванных троп','Промышляет грабежом на удалённых дорогах Мириала. Использует горные обходы и заброшенные водосборники как укрытия для награбленных припасов.','humanoid'),
 ('mirial_renegade_ascetic','Отступник горного скита','Присоединился к группе, занявшей опустевшее убежище. Под предлогом защиты своих тайн вымогает припасы у путников и преследует нежелательных свидетелей.','humanoid'),
 ('mirial_relic_guardian','Страж горного хранилища','Древний механизм сохранился в сухих внутренних камерах. Нарушение периметра пробуждает его, но старые протоколы не признают нынешних хозяев убежища.','droid'),
 ('telos_iv_cannok','Каннок','Прожорливый четвероногий зверь, чьи предки были завезены для восстановления Телоса. При нехватке пищи становится агрессивным и охотно обследует чужие стоянки.','beast'),
 ('telos_iv_salvage_raider','Налётчик старых баз','Разбирает закрытые объекты Телоса ради оборудования и оружия. Защищает найденные склады, устраивая засады в заросших служебных проходах.','humanoid'),
 ('telos_iv_ecology_droid','Сбойный дроид экологического надзора','Следит за участком восстановленных земель по устаревшим инструкциям. Ошибки опознавания заставляют его атаковать существ, которых система считает угрозой охраняемой зоне.','droid'),
 ('telos_iv_combat_droid','Законсервированный боевой дроид','Долгое время оставался в технической нише старой базы. После сбоя консервации вернулся к охране периметра, не получив новых команд и списков допуска.','droid')
ON CONFLICT (code) DO NOTHING;

INSERT INTO location_npcs(location_id,npc_template_id,code,position,level,rank)
SELECT l.id,t.id,r.code,r.position,
       LEAST(l.recommended_level+r.level_offset,(SELECT max(level) FROM levels)),r.rank
FROM (VALUES
 ('carida','cliffside_landing','carida_highland_predator','carida_arms_raider','carida_training_droid'),
 ('carida','foothill_roads','carida_highland_predator','carida_arms_raider','carida_highland_predator'),
 ('carida','supply_depot','carida_arms_raider','rogue_loader','carida_training_droid'),
 ('carida','dry_canyons','carida_highland_predator','cave_predator','carida_arms_raider'),
 ('carida','abandoned_barracks','carida_arms_raider','carida_training_droid','carida_training_droid'),
 ('carida','obstacle_range','carida_training_droid','carida_arms_raider','carida_siege_automaton'),
 ('carida','mountain_rail','carida_arms_raider','rogue_loader','carida_training_droid'),
 ('carida','sealed_armory','carida_training_droid','carida_arms_raider','carida_siege_automaton'),
 ('carida','siege_range','carida_training_droid','carida_siege_automaton','carida_siege_automaton'),
 ('carida','plateau_fortress','carida_arms_raider','carida_training_droid','carida_siege_automaton'),
 ('cato_neimoidia','arrival_terraces','cato_neimoidia_bridge_raider','contraband_guard','cato_neimoidia_vault_guard'),
 ('cato_neimoidia','cargo_bridges','cato_neimoidia_bridge_raider','rogue_loader','cato_neimoidia_vault_guard'),
 ('cato_neimoidia','lower_arcades','cato_neimoidia_bridge_raider','gang_enforcer','cato_neimoidia_vault_guard'),
 ('cato_neimoidia','misty_cliffs','cato_neimoidia_cliff_predator','cato_neimoidia_bridge_raider','cato_neimoidia_cliff_predator'),
 ('cato_neimoidia','underbridge_walkways','cato_neimoidia_bridge_raider','cato_neimoidia_cliff_predator','rogue_sentry'),
 ('cato_neimoidia','merchant_warehouses','contraband_guard','cato_neimoidia_vault_guard','cato_neimoidia_security_droid'),
 ('cato_neimoidia','silent_pavilions','cato_neimoidia_bridge_raider','cato_neimoidia_vault_guard','rogue_sentry'),
 ('cato_neimoidia','treasury_approach','cato_neimoidia_vault_guard','rogue_sentry','cato_neimoidia_security_droid'),
 ('cato_neimoidia','suspended_vaults','cato_neimoidia_vault_guard','cato_neimoidia_security_droid','cato_neimoidia_security_droid'),
 ('cato_neimoidia','sealed_treasury','cato_neimoidia_security_droid','cato_neimoidia_vault_guard','cato_neimoidia_security_droid'),
 ('shili','grassland_landing','shili_grass_stalker','shili_poacher','shili_trap_droid'),
 ('shili','colored_plains','shili_grass_stalker','shili_grass_stalker','shili_akul'),
 ('shili','forest_margin','shili_grass_stalker','shili_poacher','shili_trap_droid'),
 ('shili','river_reeds','shili_grass_stalker','shili_trap_droid','shili_akul'),
 ('shili','poacher_camps','shili_poacher','shili_trap_droid','shili_poacher'),
 ('shili','redgrass_ravines','shili_grass_stalker','shili_poacher','shili_akul'),
 ('shili','ancient_groves','shili_poacher','shili_grass_stalker','shili_akul'),
 ('shili','hidden_pens','shili_poacher','shili_trap_droid','shili_akul'),
 ('shili','akul_hunting_grounds','shili_grass_stalker','shili_akul','shili_akul'),
 ('shili','predator_plateau','shili_grass_stalker','shili_akul','shili_akul'),
 ('ord_radama','marshside_quay','ord_radama_marsh_stalker','ord_radama_dock_smuggler','ord_radama_canal_droid'),
 ('ord_radama','canal_lanes','ord_radama_dock_smuggler','gang_enforcer','ord_radama_gang_captain'),
 ('ord_radama','damp_warehouses','ord_radama_dock_smuggler','rogue_loader','ord_radama_gang_captain'),
 ('ord_radama','reed_fens','ord_radama_marsh_stalker','swamp_predator','ord_radama_marsh_stalker'),
 ('ord_radama','old_waterworks','ord_radama_canal_droid','ord_radama_dock_smuggler','ord_radama_canal_droid'),
 ('ord_radama','flood_tunnels','ord_radama_marsh_stalker','ord_radama_canal_droid','ord_radama_dock_smuggler'),
 ('ord_radama','abandoned_residences','ord_radama_dock_smuggler','gang_enforcer','ord_radama_gang_captain'),
 ('ord_radama','smugglers_islands','ord_radama_dock_smuggler','ord_radama_marsh_stalker','ord_radama_gang_captain'),
 ('ord_radama','guild_depots','contraband_guard','ord_radama_canal_droid','ord_radama_gang_captain'),
 ('ord_radama','canal_citadel','ord_radama_dock_smuggler','ord_radama_canal_droid','ord_radama_gang_captain'),
 ('mirial','cold_waystation','mirial_frost_stalker','mirial_caravan_raider','mirial_caravan_raider'),
 ('mirial','wind_scoured_plains','mirial_frost_stalker','mirial_caravan_raider','mirial_frost_stalker'),
 ('mirial','caravan_gorge','mirial_caravan_raider','mirial_frost_stalker','mirial_caravan_raider'),
 ('mirial','frost_caves','cave_predator','mirial_frost_stalker','mirial_frost_stalker'),
 ('mirial','abandoned_cisterns','mirial_caravan_raider','rogue_loader','mirial_relic_guardian'),
 ('mirial','mountain_paths','mirial_frost_stalker','mirial_caravan_raider','mirial_renegade_ascetic'),
 ('mirial','deserted_retreat','mirial_renegade_ascetic','mirial_caravan_raider','mirial_relic_guardian'),
 ('mirial','sealed_cloisters','mirial_renegade_ascetic','mirial_relic_guardian','mirial_renegade_ascetic'),
 ('mirial','cliff_sanctuaries','mirial_renegade_ascetic','mirial_relic_guardian','mirial_relic_guardian'),
 ('mirial','hidden_repository','mirial_caravan_raider','mirial_renegade_ascetic','mirial_relic_guardian'),
 ('telos_iv','lakeside_landing','telos_iv_cannok','telos_iv_salvage_raider','telos_iv_ecology_droid'),
 ('telos_iv','restored_woodlands','telos_iv_cannok','telos_iv_ecology_droid','telos_iv_cannok'),
 ('telos_iv','grassy_ravines','telos_iv_cannok','telos_iv_salvage_raider','telos_iv_cannok'),
 ('telos_iv','survey_station','telos_iv_salvage_raider','telos_iv_ecology_droid','telos_iv_ecology_droid'),
 ('telos_iv','reservoir_shore','telos_iv_cannok','telos_iv_salvage_raider','telos_iv_ecology_droid'),
 ('telos_iv','overgrown_access','telos_iv_cannok','telos_iv_salvage_raider','telos_iv_combat_droid'),
 ('telos_iv','deserted_service_hub','telos_iv_salvage_raider','rogue_loader','telos_iv_ecology_droid'),
 ('telos_iv','buried_base','telos_iv_salvage_raider','telos_iv_ecology_droid','telos_iv_combat_droid'),
 ('telos_iv','control_bunkers','telos_iv_ecology_droid','telos_iv_salvage_raider','telos_iv_combat_droid'),
 ('telos_iv','sealed_defense_complex','telos_iv_combat_droid','telos_iv_ecology_droid','telos_iv_combat_droid')
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
WHERE p.code IN ('carida','cato_neimoidia','shili','ord_radama','mirial','telos_iv') AND n.code IN ('normal','strong','elite')
ON CONFLICT (location_npc_id,slot) DO NOTHING;
