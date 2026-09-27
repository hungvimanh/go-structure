using System.Reflection;

namespace CHECKIN.Common
{
    public class FilterEntity
    {
        public int Skip { get; set; }
        public int Take { get; set; }
        public string Search { get; set; }
        public OrderType OrderType { get; set; }

        public FilterEntity()
        {
            Skip = 0;
            Take = int.MaxValue;
            OrderType = OrderType.DESC;
        }

        public void TrimString()
        {
            List<PropertyInfo> PropertyInfoes = this.GetType().GetProperties().ToList();
            foreach (PropertyInfo PropertyInfo in PropertyInfoes)
            {
                if (PropertyInfo.PropertyType.IsGenericType && PropertyInfo.PropertyType.GetGenericTypeDefinition() == typeof(List<>))
                {
                    if (PropertyInfo.GetMethod != null)
                    {
                        IEnumerable<FilterEntity> FilterEntities = PropertyInfo.GetValue(this) as IEnumerable<FilterEntity>;
                        if (FilterEntities != null)
                            foreach (FilterEntity FilterEntity in FilterEntities)
                            {
                                FilterEntity.TrimString();
                            }
                    }
                }
                if (PropertyInfo.PropertyType.IsSubclassOf(typeof(FilterEntity)))
                {
                    FilterEntity FilterEntity = PropertyInfo.GetValue(this) as FilterEntity;
                    if (FilterEntity != null)
                    {
                        FilterEntity.TrimString();
                    }
                }
                if (PropertyInfo.PropertyType.Name == typeof(string).Name)
                {
                    if (PropertyInfo.GetMethod != null && PropertyInfo.SetMethod != null)
                    {
                        string value = PropertyInfo.GetValue(this) as string;
                        value = value?.Trim();
                        PropertyInfo.SetValue(this, value);
                    }
                }
            }
        }

        public void Initialize()
        {
            var properties = this.GetType().GetProperties();
            foreach (var p in properties)
            {
                if (p.PropertyType.Name == nameof(IdFilter))
                    p.SetValue(this, new IdFilter());
                if (p.PropertyType.Name == nameof(StringFilter))
                    p.SetValue(this, new StringFilter());
                if (p.PropertyType.Name == nameof(DateFilter))
                    p.SetValue(this, new DateFilter());
                if (p.PropertyType.Name == nameof(LongFilter))
                    p.SetValue(this, new LongFilter());
                if (p.PropertyType.Name == nameof(DecimalFilter))
                    p.SetValue(this, new DecimalFilter());
                if (p.PropertyType.Name == nameof(GuidFilter))
                    p.SetValue(this, new GuidFilter());
            }
        }
    }

    [JsonConverter(typeof(StringEnumConverter))]
    public enum OrderType
    {
        DESC,
        ASC,
    }
}
