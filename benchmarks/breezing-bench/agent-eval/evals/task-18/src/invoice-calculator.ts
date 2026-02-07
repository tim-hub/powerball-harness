import type { LineItem, Invoice, IInvoiceCalculator } from './types';

export class InvoiceCalculator implements IInvoiceCalculator {
  private invoices = new Map<string, Invoice>();

  createInvoice(id: string, items: LineItem[], taxRate: number): Invoice {
    const subtotal = items.reduce((sum, item) => sum + item.quantity * item.unitPrice, 0);
    const tax = subtotal * taxRate;
    const total = subtotal + tax;

    const invoice: Invoice = { id, items: [...items], taxRate, subtotal, tax, total };
    this.invoices.set(id, invoice);
    return invoice;
  }

  getInvoice(id: string): Invoice | undefined {
    return this.invoices.get(id);
  }

  addItem(invoiceId: string, item: LineItem): Invoice | undefined {
    const invoice = this.invoices.get(invoiceId);
    if (!invoice) return undefined;

    invoice.items.push(item);
    invoice.subtotal = invoice.items.reduce((sum, i) => sum + i.quantity * i.unitPrice, 0);
    invoice.tax = invoice.subtotal * invoice.taxRate;
    invoice.total = invoice.subtotal + invoice.tax;

    return invoice;
  }
}
