export function processData(data: any): any {
  if (!data) return null;

  const result: any = {};

  if (data.type === 'user') {
    result.name = data.payload.name;
    result.age = data.payload.age;
    result.email = data.payload.email;
    result.tags = data.payload.tags || [];
  } else if (data.type === 'product') {
    result.title = data.payload.title;
    result.price = data.payload.price;
    result.category = data.payload.category;
    result.inStock = data.payload.inStock ?? true;
  } else if (data.type === 'order') {
    result.orderId = data.payload.orderId;
    result.items = data.payload.items.map((item: any) => ({
      productId: item.id,
      quantity: item.qty,
      subtotal: item.price * item.qty,
    }));
    result.total = result.items.reduce((sum: any, item: any) => sum + item.subtotal, 0);
  }

  return result;
}

export function validateInput(input: any): any {
  const errors: any[] = [];

  if (!input.type) errors.push({ field: 'type', message: 'required' });
  if (!input.payload) errors.push({ field: 'payload', message: 'required' });

  if (input.type === 'user') {
    if (!input.payload?.name) errors.push({ field: 'name', message: 'required' });
    if (!input.payload?.email) errors.push({ field: 'email', message: 'required' });
  }

  return { valid: errors.length === 0, errors };
}

export function transformBatch(items: any[]): any[] {
  return items.map((item: any) => processData(item)).filter((r: any) => r !== null);
}
