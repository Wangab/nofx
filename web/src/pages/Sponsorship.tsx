import { useState } from 'react'
import { Heart, Copy, Check } from 'lucide-react'
import { t, type Language } from '../i18n/translations'

interface SponsorshipProps {
  language?: Language
}

export default function Sponsorship({ language = 'zh' as Language }: SponsorshipProps) {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  const sponsorshipMethods = [
    // {
    //   name: language === 'zh' ? '支付宝' : 'Alipay',
    //   icon: '💳',
    //   address: 'your-alipay-account@example.com',
    //   qrCode: '/images/alipay-qr.png', // Replace with actual QR code
    // },
    // {
    //   name: language === 'zh' ? '微信支付' : 'WeChat Pay',
    //   icon: '💚',
    //   address: 'your-wechat-id',
    //   qrCode: '/images/wechat-qr.png', // Replace with actual QR code
    // },
    // {
    //   name: language === 'zh' ? 'PayPal' : 'PayPal',
    //   icon: '🅿️',
    //   address: 'your-paypal@example.com',
    //   qrCode: '/images/paypal-qr.png', // Replace with actual QR code
    // },
    // {
    //   name: language === 'zh' ? 'TRC20-USDT' : 'Crypto',
    //   icon: '₿',
    //   address: '1A1z7agoat3Z2LaFBhPP7P2p5H9pZ7Y8m',
    //   qrCode: '/images/crypto-qr.png', // Replace with actual QR code
    // },
    {
      name: language === 'zh' ? 'TRC20-USDT' : 'TRC20-USDT',
      icon: '₿',
      address: 'TXRc2qLsEQKzBN81B2a4k5YrfK96EuptUn',
      qrCode: '/images/crypto-usdt-qr.jpg', // Replace with actual QR code
    },
    {
      name: language === 'zh' ? 'ERC20-ETH' : 'ERC20-ETH',
      icon: 'E',
      address: '0xF2ebF6bE069d94b863670455A05F5b7929eF624f',
      qrCode: '/images/crypto-eth-qr.jpg', // Replace with actual QR code
    },
  ]

  const handleCopyAddress = (address: string, index: number) => {
    navigator.clipboard.writeText(address)
    setCopiedIndex(index)
    setTimeout(() => setCopiedIndex(null), 2000)
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-nofx-bg-dark via-nofx-bg-dark to-nofx-bg-lighter pt-32 pb-16">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="text-center mb-16">
          <div className="flex items-center justify-center gap-3 mb-4">
            <Heart className="w-8 h-8 text-nofx-gold" fill="currentColor" />
            <h1 className="text-4xl sm:text-5xl font-bold text-white">
              {t('sponsorshipTitle', language)}
            </h1>
          </div>
          <p className="text-xl text-nofx-text-muted max-w-2xl mx-auto">
            {t('sponsorshipDesc', language)}
          </p>
        </div>

        {/* Sponsorship Methods Grid */}
        <div className="mb-16">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 w-fit mx-auto">
            {sponsorshipMethods.map((method, index) => (
              <div
                key={index}
                className="rounded-xl bg-nofx-bg-lighter border border-nofx-gold/20 hover:border-nofx-gold/50 transition-all duration-300 overflow-hidden group"
              >
                {/* Card Content */}
                <div className="p-6 flex flex-col h-full">
                  {/* Icon and Name */}
                  <div className="flex items-center gap-3 mb-6">
                    {/* <span className="text-3xl">{method.icon}</span> */}
                    <h3 className="text-lg font-bold text-white">{method.name}</h3>
                  </div>

                  {/* QR Code */}
                  <div className="bg-white/5 rounded-lg p-4 mb-6 flex items-center justify-center h-40 border border-nofx-gold/10">
                    <img 
                      src={method.qrCode} 
                      alt={`${method.name} QR Code`}
                      className="w-full h-full object-contain"
                      onError={(e) => {
                        // Fallback if image fails to load
                        (e.target as HTMLImageElement).style.display = 'none'
                      }}
                    />
                  </div>

                  {/* Address */}
                  <div className="mb-4 flex-grow">
                    <p className="text-xs text-nofx-text-muted mb-2">
                      {language === 'zh' ? '地址' : 'Address'}:
                    </p>
                    <p className="text-xs font-mono text-white break-all bg-black/20 rounded p-2">
                      {method.address}
                    </p>
                  </div>

                  {/* Copy Button */}
                  <button
                    onClick={() => handleCopyAddress(method.address, index)}
                    className="w-full py-2 px-3 rounded-lg bg-nofx-gold/10 hover:bg-nofx-gold/20 text-nofx-gold transition-colors flex items-center justify-center gap-2 text-sm font-semibold"
                  >
                    {copiedIndex === index ? (
                      <>
                        <Check className="w-4 h-4" />
                        {language === 'zh' ? '已复制' : 'Copied'}
                      </>
                    ) : (
                      <>
                        <Copy className="w-4 h-4" />
                        {language === 'zh' ? '复制地址' : 'Copy Address'}
                      </>
                    )}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Appreciation Section */}
        <div className="text-center bg-nofx-bg-lighter border border-nofx-gold/20 rounded-xl p-8 md:p-12">
          <h2 className="text-2xl font-bold text-white mb-4">
            {language === 'zh' ? '感谢您的支持' : 'Thank You For Your Support'}
          </h2>
          <p className="text-nofx-text-muted text-lg mb-8 max-w-3xl mx-auto">
            {language === 'zh'
              ? '无论捐赠多少，您的每一份支持都对我们意义重大。这些资金将用于平台的持续开发、改进功能和提供更好的用户体验。'
              : 'Every contribution, no matter the amount, means a lot to us. Your support helps us continue developing the platform, improving features, and providing a better user experience.'}
          </p>

          {/* Sponsors List */}
          <div className="text-center">
            <p className="text-sm text-nofx-text-muted mb-4">
              {language === 'zh' ? '感谢所有赞助者' : 'Thank you to all our supporters'}
            </p>
            <div className="flex flex-wrap justify-center gap-6">
              {(() => {
                const sponsors = [
                  { name: 'Alex Chen', initials: 'AC' },
                  { name: 'Sarah Wilson', initials: 'SW' },
                  { name: '张明', initials: 'ZM' },
                  { name: 'James Liu', initials: 'JL' },
                  { name: '李雨晴', initials: 'LY' },
                  { name: 'Michael Zhang', initials: 'MZ' },
                  { name: 'Emma Davis', initials: 'ED' },
                  { name: '王健', initials: 'WJ' },
                ]
                
                return sponsors.map((sponsor, i) => (
                  <div
                    key={i}
                    className="flex flex-col items-center group cursor-pointer"
                  >
                    <div className="w-14 h-14 rounded-full bg-gradient-to-br from-nofx-gold/30 to-nofx-gold/10 border border-nofx-gold/30 flex items-center justify-center group-hover:border-nofx-gold/60 transition-all duration-300 mb-2">
                      <span className="text-sm font-bold text-nofx-gold">
                        {sponsor.initials}
                      </span>
                    </div>
                    <span className="text-xs text-nofx-text-muted group-hover:text-nofx-gold transition-colors">
                      {sponsor.name}
                    </span>
                  </div>
                ))
              })()}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}