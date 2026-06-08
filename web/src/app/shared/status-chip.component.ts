import { Component, Input } from '@angular/core';
import { ShopStatus, STATUS_META } from '../core/models/shop.model';

/** Mali prikaz statusa prodavnice (Provisioning / Ready / Error...). */
@Component({
  selector: 'app-status-chip',
  standalone: true,
  template: `
    <span [class]="'status-chip ' + meta.css" [attr.aria-label]="'Status: ' + meta.label">
      <span class="dot"></span>{{ meta.label }}
    </span>
  `,
})
export class StatusChipComponent {
  @Input({ required: true }) status!: ShopStatus;

  get meta() {
    return STATUS_META[this.status] ?? STATUS_META.pending;
  }
}
