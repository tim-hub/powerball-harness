export interface LineItem {
  description: string;
  quantity: number;
  unitPrice: number;
}

export interface Invoice {
  id: string;
  items: LineItem[];
  taxRate: number;  // e.g. 0.10 for 10%
  subtotal: number;
  tax: number;
  total: number;
  discount?: number;  // percentage (0-100)
}

export interface IInvoiceCalculator {
  createInvoice(id: string, items: LineItem[], taxRate: number): Invoice;
  getInvoice(id: string): Invoice | undefined;
  addItem(invoiceId: string, item: LineItem): Invoice | undefined;
  applyDiscount(invoiceId: string, discountPercent: number): Invoice | undefined;
}
