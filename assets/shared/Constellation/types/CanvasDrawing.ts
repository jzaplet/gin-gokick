export type CanvasDrawing = Pick<
    CanvasRenderingContext2D,
    'arc' | 'beginPath' | 'clearRect' | 'fill' | 'fillStyle' | 'lineTo' | 'lineWidth' | 'moveTo'
    | 'setTransform' | 'stroke' | 'strokeStyle'
>;
