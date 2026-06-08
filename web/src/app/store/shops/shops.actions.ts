import { createActionGroup, emptyProps, props } from '@ngrx/store';
import {
  Shop,
  CreateShopPayload,
  UpdateShopPayload,
} from '../../core/models/shop.model';

export const ShopsActions = createActionGroup({
  source: 'Shops',
  events: {
    Load: emptyProps(),
    'Load Success': props<{ shops: Shop[] }>(),
    'Load Failure': props<{ error: string }>(),

    'Load One': props<{ id: string }>(),
    'Load One Success': props<{ shop: Shop }>(),
    'Load One Failure': props<{ error: string }>(),

    Create: props<{ payload: CreateShopPayload }>(),
    'Create Success': props<{ shop: Shop }>(),
    'Create Failure': props<{ error: string }>(),

    Update: props<{ id: string; patch: UpdateShopPayload }>(),
    'Update Success': props<{ shop: Shop }>(),
    'Update Failure': props<{ error: string }>(),

    Delete: props<{ id: string; name: string }>(),
    'Delete Success': props<{ id: string }>(),
    'Delete Failure': props<{ error: string }>(),
  },
});
