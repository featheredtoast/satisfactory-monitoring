Map projections in grafana are not perfect. Grafana does not currently yet have support for flat game maps, such as satisfactory.

I'm faking the coordinates on the returned satisfactory-calculator xyz slippy tile map into the projected world coordinates. Building placements may not be perfect.

using satisfactory-calculator interactive map tiles for map:
`https://static.satisfactory-calculator.com/imgMap/gameLayer/EarlyAccess/{z}/{x}/{y}.png?v=1671697795`
world coordinates
-157.54585, 82.66079
22.45923, -21.97246
180.004, 104.632(world)
7499, 7500 (game)
ratio: 0.024004, 0.013951

Note: due to projections, I've nudged the latitude offset a bit higher as the offsets are not great with the given calculations. Nudged to: +92.66079 rather than 82.66079.

calculations are calculated as:
in-game x = internal-x / 100
in-game y = internal-y / 100
longitude = ((internal-x + 324600) * 0.00024004) - 157.54585
latitude = ((internal-y + 375000) * -0.00013951) + 92.66079
(Latitude is flipped from in game coordinate system)
