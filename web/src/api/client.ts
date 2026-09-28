import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      console.error('Unauthorized')
    } else if (error.response?.status >= 500) {
      console.error('Server error:', error.message)
    }
    return Promise.reject(error)
  }
)

export interface PayChannel {
  ALIPAY: 'ALIPAY'
  WECHAT: 'WECHAT'
  UNIONPAY: 'UNIONPAY'
  CRYPTO: 'CRYPTO'
  BANK_CARD: 'BANK_CARD'
}

export interface OrderStatus {
  PENDING: 'PENDING'
  PAID: 'PAID'
  SETTLED: 'SETTLED'
  FAILED: 'FAILED'
  CLOSED: 'CLOSED'
  REFUNDED: 'REFUNDED'
}

export interface Order {
  id: number
  merchant_id: string
  out_order_no: string
  subject: string
  amount_cny: number
  channel: string
  status: string
  qr_code_url?: string
  pay_url?: string
  trade_no?: string
  transaction_id?: string
  paid_at?: string
  settled_at?: string
  refunded_at?: string
  created_at: string
  updated_at: string
}

export interface CreateOrderRequest {
  merchant_id: string
  out_order_no: string
  subject: string
  amount_cny: number
  channel: string
  wallet_addr?: string
}

export interface QRPaymentRequest {
  merchant_id: string
  out_order_no: string
  subject: string
  amount_cny: number
  channel: string
  wallet_addr?: string
  expire_minutes?: number
  callback_url?: string
}

export interface QRPaymentResponse {
  order_id: number
  out_order_no: string
  qr_code_url: string
  qr_code_base64?: string
  channel: string
  amount_cny: number
  status: string
  expires_at: string
  created_at: string
}

export interface PaymentStatusResponse {
  order_id: number
  out_order_no: string
  status: string
  amount_cny: number
  channel: string
  created_at: string
  paid_at?: string
  trade_no?: string
  transaction_id?: string
  qr_code_url?: string
  expires_at: string
}

export interface AggregationPaymentRequest {
  merchant_id: string
  out_order_no: string
  subject: string
  amount_cny: number
  expire_minutes?: number
  callback_url?: string
}

export interface ChannelInfo {
  channel: string
  qr_code_url?: string
}

export interface AggregationPaymentResponse {
  order_id: number
  out_order_no: string
  qr_code_url: string
  channels: ChannelInfo[]
  status: string
  expires_at: string
  created_at: string
}

export interface WebhookEvent {
  order_id?: number
  out_order_no: string
  status: string
  amount_cny: number
  trade_no?: string
  paid_at?: string
  timestamp: number
  nonce: string
}

export interface CryptoRedemption {
  id: number
  order_id?: number
  wallet_address: string
  crypto_asset: string
  crypto_amount: number
  fiat_cny: number
  exchange_rate: number
  tx_hash?: string
  network?: string
  status: string
  created_at: string
  confirmed_at?: string
}

export interface ExchangeQuoteRequest {
  asset: string
  fiat_cny: number
  order_id?: string
}

export interface ExchangeRedeemRequest {
  order_id: string
  wallet_addr: string
  asset?: string
  network?: string
}

export const paymentApi = {
  createOrder(data: CreateOrderRequest): Promise<{ data: Order }> {
    return api.post('/pay', data)
  },

  createQRCodePayment(data: QRPaymentRequest): Promise<{ data: QRPaymentResponse }> {
    return api.post('/qrcode', data)
  },

  queryPayment(outOrderNo: string): Promise<{ data: PaymentStatusResponse }> {
    return api.get('/qrcode', { params: { out_order_no: outOrderNo } })
  },

  cancelPayment(outOrderNo: string): Promise<{ data: any }> {
    return api.put('/qrcode', {}, { params: { out_order_no: outOrderNo } })
  },

  createAggregationPayment(data: AggregationPaymentRequest): Promise<{ data: AggregationPaymentResponse }> {
    return api.post('/qrcode/aggregation', data)
  },

  resolveAggregation(outOrderNo: string, channel?: string): Promise<{ data: any }> {
    const params: any = { out_order_no: outOrderNo }
    if (channel) params.channel = channel
    return api.get('/qrcode/resolve', { params })
  },

  listOrders(merchantId?: string): Promise<{ data: { orders: Order[]; count: number } }> {
    return api.get('/orders', { params: merchantId ? { merchant_id: merchantId } : {} })
  },

  getOrder(id: number): Promise<{ data: Order }> {
    return api.get(`/orders/${id}`)
  },

  sendWebhook(event: WebhookEvent, signature?: string): Promise<{ data: any }> {
    const headers: any = {}
    if (signature) headers['X-Signature'] = signature
    return api.post('/webhook', event, { headers })
  },

  cryptoQuote(data: ExchangeQuoteRequest): Promise<{ data: CryptoRedemption }> {
    return api.post('/exchange/quote', data)
  },

  cryptoRedeem(data: ExchangeRedeemRequest): Promise<{ data: CryptoRedemption }> {
    return api.post('/exchange/redeem', data)
  },

  getRedemption(orderId: number): Promise<{ data: CryptoRedemption }> {
    return api.get('/exchange', { params: { order_id: orderId } })
  },

  healthCheck(): Promise<{ data: any }> {
    return axios.get('/health')
  },
}

export default api
