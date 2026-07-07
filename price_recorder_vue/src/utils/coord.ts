import gcoord from 'gcoord'

/**
 * Convert GCJ-02 (AMap/Gaode) coordinates to WGS-84 (GPS) coordinates.
 * AMap events (rightclick, dragend) return GCJ-02; the app stores RAW
 * WGS-84 per CR-2.
 */
export function gcj02ToWgs84(
  gcj: [number, number],
): { lng: number; lat: number } {
  const [lng, lat] = gcoord.transform(gcj, gcoord.GCJ02, gcoord.WGS84)
  return { lng, lat }
}

/**
 * Synchronous WGS-84 → GCJ-02 conversion for map operations (centering,
 * marker placement). Unlike AMap.convertFrom (async, network), this
 * uses the local gcoord transform and never blocks.
 */
export function wgs84ToGcj02(
  lng: number,
  lat: number,
): [number, number] {
  return gcoord.transform([lng, lat], gcoord.WGS84, gcoord.GCJ02) as [number, number]
}
